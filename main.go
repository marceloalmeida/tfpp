package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Platform struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

type Version struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

type Versions struct {
	Versions []Version `json:"versions"`
}

type WellKnown struct {
	ProvidersV1 string `json:"providers.v1"`
	ModulesV1   string `json:"modules.v1"`
}

var defaultWellKnownData = WellKnown{
	ProvidersV1: "/v1/providers/",
	ModulesV1:   "/v1/modules/",
}

type Architecture struct {
	Protocols           []string `json:"protocols"`
	Os                  string   `json:"os"`
	Arch                string   `json:"arch"`
	Filename            string   `json:"filename"`
	DownloadUrl         string   `json:"download_url"`
	ShasumsUrl          string   `json:"shasums_url"`
	ShasumsSignatureUrl string   `json:"shasums_signature_url"`
	Shasum              string   `json:"shasum"`
	SigningKeys         struct {
		GpgPublicKeys []struct {
			KeyId          string `json:"key_id"`
			AsciiArmor     string `json:"ascii_armor"`
			TrustSignature string `json:"trust_signature"`
			Source         string `json:"source"`
			SourceUrl      string `json:"source_url"`
		} `json:"gpg_public_keys"`
	} `json:"signing_keys"`
}

func main() {
	log.Println("📦 Packaging Terraform Provider for private registry...")

	namespace := flag.String("ns", "", "Namespace for the Terraform registry.")
	domain := flag.String("d", "", "Private Terraform registry domain.")
	providerName := flag.String("p", "", "Name of the Terraform provider.")
	distPath := flag.String("dp", "dist", "Path to Go Releaser build files.")
	repoName := flag.String("r", "", "Name of the provider repository used in Go Releaser build name.")
	version := flag.String("v", "", "Semantic version of build.")
	gpgFingerprint := flag.String("gf", "", "GPG Fingerprint of key used by Go Releaser")
	gpgPubKeyFile := flag.String("gk", "pubkey.txt", "Path to GPG Public Key in ASCII Armor format.")
	flag.Parse()

	if *namespace == "" {
		log.Fatal("Namespace is required.")
	}
	if *domain == "" {
		log.Fatal("Domain is required.")
	}
	if *providerName == "" {
		log.Fatal("Provider name is required.")
	}
	if *repoName == "" {
		log.Fatal("Repository name is required.")
	}
	if *version == "" {
		log.Fatal("Version is required.")
	}
	if *gpgFingerprint == "" {
		log.Fatal("GPG Fingerprint is required.")
	}

	*distPath = *distPath + "/"

	err := deleteDir("release")
	if err != nil {
		log.Fatalf("Error deleting 'release' dir: %s", err)
	}

	err = createDir("release")
	if err != nil {
		log.Fatalf("Error creating 'release' dir: %s", err)
	}

	err = provider(*namespace, *providerName, *distPath, *repoName, *version, *gpgFingerprint, *gpgPubKeyFile, *domain)
	if err != nil {
		log.Fatalf("Error creating dir: %s", err)
	}

	log.Println("🎉 Packaged Terraform Provider for private registry.")
}

func provider(namespace, provider, distPath, repoName, version, gpgFingerprint, gpgPubKeyFile, domain string) error {
	wellKnownData, err := createVersionsFile(namespace, provider, distPath, repoName, version, domain)
	if err != nil {
		log.Fatalf("Error creating versions file: %s", err)
	}

	versionPath := providerDirs(namespace, provider, version, wellKnownData)

	copyShaFiles(versionPath, distPath, repoName, version)
	downloadPath, err := createDownloadsDir(versionPath)
	if err != nil {
		return err
	}

	err = createTargetDirs(*downloadPath)
	if err != nil {
		log.Fatalf("Error creating target dirs: %s", err)
	}

	err = copyBuildZips(*downloadPath, distPath, repoName, version)
	if err != nil {
		log.Fatalf("Error copying build zips: %s", err)
	}

	err = createArchitectureFiles(namespace, provider, distPath, repoName, version, gpgFingerprint, gpgPubKeyFile, domain, wellKnownData)
	if err != nil {
		log.Fatalf("Error creating architecture files: %s", err)
	}

	return nil
}

func providerDirs(namespace, repoName, version string, wewellKnownData WellKnown) string {
	log.Println("* Creating release/[well know providers]/[namespace]/[repo]/[version] directories")

	providerPathArr := [5]string{"release", wewellKnownData.ProvidersV1, namespace, repoName, version}

	var currentPath string
	for _, v := range providerPathArr {
		currentPath = currentPath + v + "/"
		err := createDir(currentPath)
		if err != nil {
			log.Fatalf("Error creating directory: %s", err)
		}
	}

	return currentPath
}

func createVersionsFile(namespace, provider, distPath, repoName, version, domain string) (WellKnown, error) {
	registryVersionFile, wellKnownData, _ := downloadVersionsFile(namespace, provider, domain)

	versionPath := fmt.Sprintf("%s%s%s/%s/versions", "release", wellKnownData.ProvidersV1, namespace, provider)

	shaSumContents, err := getShaSumContents(distPath, repoName, version)
	if err != nil {
		return wellKnownData, err
	}

	var ver Version
	ver.Version = version
	ver.Protocols = []string{"4.0", "5.1"}
	ver.Platforms = []Platform{}

	var vers Versions
	vers.Versions = []Version{}

	for _, line := range shaSumContents {
		fileName := line[1]

		removeFileExtension := strings.Split(fileName, ".zip")
		fileNameSplit := strings.Split(removeFileExtension[0], "_")

		if len(fileNameSplit) < 4 {
			log.Printf("Filename '%s' is not in the expected format, skipping...", fileName)
			continue
		}

		target := fileNameSplit[2]
		arch := fileNameSplit[3]

		var plat Platform
		plat.Os = target
		plat.Arch = arch

		ver.Platforms = append(ver.Platforms, plat)
	}

	for _, v := range registryVersionFile.Versions {
		exists := false
		for _, existingVersion := range vers.Versions {
			if existingVersion.Version == v.Version {
				exists = true
				break
			}
		}
		if !exists {
			vers.Versions = append(vers.Versions, v)
		}
	}

	vers.Versions = append(vers.Versions, ver)

	versionsFile, err := json.MarshalIndent(vers, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %s", err)
	}

	versionPathDir := strings.Split(versionPath, "/")
	err = createDirRecursive(strings.Join(versionPathDir[:len(versionPathDir)-1], "/"))
	if err != nil {
		log.Fatalf("Error creating directory: %s", err)
		return wellKnownData, err
	}

	err = writeFile(versionPath, versionsFile)
	if err != nil {
		log.Fatalf("Error writing versions file: %s", err)
		return wellKnownData, err
	}

	return wellKnownData, nil
}

func downloadVersionsFile(namespace, provider, domain string) (Versions, WellKnown, error) {
	log.Println("* Downloading versions file")

	httpClient := &http.Client{}
	wellKnownUrl := fmt.Sprintf("https://%s/.well-known/terraform.json", domain)
	resp, err := httpClient.Get(wellKnownUrl)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Error downloading well-known file: %s or status code not %d", err.Error(), http.StatusOK)

		wellKnownFile, err := json.MarshalIndent(defaultWellKnownData, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %s", err)
		}
		err = createDirRecursive("release/.well-known/")
		if err != nil {
			log.Fatalf("Error creating directory: %s", err)
		}
		err = writeFile("release/"+".well-known/terraform.json", wellKnownFile)
		if err != nil {
			log.Fatalf("Error writing terraform.json file: %s", err)
		}

		return Versions{}, defaultWellKnownData, fmt.Errorf("error downloading well-known file: %s or status code not %d", err.Error(), http.StatusOK)
	}
	defer resp.Body.Close()

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return Versions{}, defaultWellKnownData, err
	}
	resp.Body.Close()

	var wellKnownData WellKnown
	err = json.Unmarshal(bodyResp, &wellKnownData)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %s", err)
		return Versions{}, defaultWellKnownData, err
	}

	versionsUrl := fmt.Sprintf("https://%s%s%s/%s/versions", domain, wellKnownData.ProvidersV1, namespace, provider)
	resp, err = httpClient.Get(versionsUrl)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Error downloading versions file: %s or status code not %d", err, http.StatusOK)
		return Versions{}, wellKnownData, fmt.Errorf("error downloading versions file: %s or status code not %d", err, http.StatusOK)
	}
	defer resp.Body.Close()

	bodyResp, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return Versions{}, wellKnownData, err
	}
	resp.Body.Close()
	var versionsData Versions
	err = json.Unmarshal(bodyResp, &versionsData)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %s", err)
		return Versions{}, wellKnownData, err
	}

	return versionsData, wellKnownData, nil
}

func copyShaFiles(destPath, srcPath, repoName, version string) {
	log.Printf("* Copying SHA files in %s directory", srcPath)

	shaSum := repoName + "_" + version + "_SHA256SUMS"
	shaSumPath := srcPath + "/" + shaSum

	err := copyFile(shaSumPath, destPath+shaSum)
	if err != nil {
		log.Fatal(err)
	}

	err = copyFile(shaSumPath+".sig", destPath+shaSum+".sig")
	if err != nil {
		log.Fatal(err)
	}
}

func createDownloadsDir(destPath string) (*string, error) {
	log.Printf("* Creating download/ in %s directory", destPath)

	downloadPath := destPath + "download/"

	err := createDir(downloadPath)
	if err != nil {
		return nil, err
	}

	return &downloadPath, nil
}

func createTargetDirs(destPath string) error {
	log.Printf("* Creating target dirs in %s directory", destPath)

	targets := [4]string{"darwin", "freebsd", "linux", "windows"}

	for _, v := range targets {
		err := createDir(destPath + v)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyBuildZips(destPath, distPath, repoName, version string) error {
	log.Println("* Copying build zips")

	shaSumContents, err := getShaSumContents(distPath, repoName, version)
	if err != nil {
		return err
	}

	for _, v := range shaSumContents {
		zipName := v[1]

		if !strings.HasSuffix(zipName, ".zip") {
			log.Printf("Filename '%s' is not a zip file, skipping...", zipName)
			continue
		}

		zipSrcPath := distPath + zipName
		zipDestPath := destPath + zipName

		log.Printf("  - Zip Source: %s", zipSrcPath)
		log.Printf("   - Zip Dest:  %s", zipDestPath)

		err := copyFile(zipSrcPath, zipDestPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func getShaSumContents(distPath, repoName, version string) ([][]string, error) {
	shaSumFileName := repoName + "_" + version + "_SHA256SUMS"
	shaSumPath := distPath + "/" + shaSumFileName

	shaSumLine, err := readFile(shaSumPath)
	if err != nil {
		return nil, err
	}

	buildsAndShaSums := [][]string{}

	for _, line := range shaSumLine {
		lineSplit := strings.Split(line, "  ")

		row := []string{lineSplit[0], lineSplit[1]}
		buildsAndShaSums = append(buildsAndShaSums, row)
	}

	return buildsAndShaSums, nil
}

func createArchitectureFiles(namespace, provider, distPath, repoName, version, gpgFingerprint, gpgPubKeyFile, domain string, wellKnownData WellKnown) error {
	log.Println("* Creating architecture files in target directories")

	prefix := fmt.Sprintf("%s%s/%s/%s/", wellKnownData.ProvidersV1, namespace, provider, version)
	pathPrefix := fmt.Sprintf("release%s", prefix)
	urlPrefix := fmt.Sprintf("https://%s%s", domain, prefix)

	downloadUrlPrefix := urlPrefix + "download/"
	downloadPathPrefix := pathPrefix + "download/"

	shasumsUrl := urlPrefix + fmt.Sprintf("%s_%s_SHA256SUMS", repoName, version)
	shasumsSigUrl := shasumsUrl + ".sig"

	shaSumContents, err := getShaSumContents(distPath, repoName, version)
	if err != nil {
		return err
	}

	gpgFile, err := readFile(gpgPubKeyFile)
	if err != nil {
		log.Fatalf("Error reading '%s' file: %s", gpgPubKeyFile, err)
	}

	gpgAsciiPub := ""
	for _, line := range gpgFile {
		gpgAsciiPub = gpgAsciiPub + line + "\n"
	}

	for _, line := range shaSumContents {
		shasum := line[0]
		fileName := line[1]

		downloadUrl := downloadUrlPrefix + fileName

		removeFileExtension := strings.Split(fileName, ".zip")
		fileNameSplit := strings.Split(removeFileExtension[0], "_")

		if len(fileNameSplit) < 4 {
			log.Printf("Filename '%s' is not in the expected format, skipping...", fileName)
			continue
		}

		target := fileNameSplit[2]
		arch := fileNameSplit[3]

		archFileName := downloadPathPrefix + target + "/" + arch

		var architecture Architecture
		architecture.Protocols = []string{"4.0", "5.1"}
		architecture.Os = target
		architecture.Arch = arch
		architecture.Filename = fileName
		architecture.DownloadUrl = downloadUrl
		architecture.ShasumsUrl = shasumsUrl
		architecture.ShasumsSignatureUrl = shasumsSigUrl
		architecture.Shasum = shasum
		architecture.SigningKeys.GpgPublicKeys = []struct {
			KeyId          string `json:"key_id"`
			AsciiArmor     string `json:"ascii_armor"`
			TrustSignature string `json:"trust_signature"`
			Source         string `json:"source"`
			SourceUrl      string `json:"source_url"`
		}{
			{
				KeyId:          gpgFingerprint,
				AsciiArmor:     gpgAsciiPub,
				TrustSignature: "",
				Source:         "",
				SourceUrl:      "",
			},
		}
		architectureTemplate, err := json.MarshalIndent(architecture, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %s", err)
		}

		log.Printf("  - Arch file: %s", archFileName)

		err = writeFile(archFileName, architectureTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}

func createDir(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}
	err := os.Mkdir(path, os.ModePerm)
	return err
}

func createDirRecursive(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func readFile(filePath string) ([]string, error) {
	readFile, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	readFile.Close()

	return fileLines, nil
}

func writeFile(fileName string, fileContents []byte) error {
	err := os.WriteFile(fileName, fileContents, 0644)
	return err
}
