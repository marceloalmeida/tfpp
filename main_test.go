package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateDir tests the createDir function.
func TestCreateDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    "/tmp/test_create_dir",
			wantErr: false,
		},
		{
			name:    "create existing directory",
			path:    "/tmp/test_create_dir_existing",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.RemoveAll(tt.path)
			defer os.RemoveAll(tt.path)

			// For existing directory test, create it first
			if strings.Contains(tt.name, "existing") {
				err := os.Mkdir(tt.path, os.ModePerm)
				if err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			err := createDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify directory exists
			if _, err := os.Stat(tt.path); os.IsNotExist(err) && !tt.wantErr {
				t.Errorf("Directory was not created: %v", tt.path)
			}
		})
	}
}

// TestCreateDirRecursive tests the createDirRecursive function.
func TestCreateDirRecursive(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create nested directories",
			path:    "/tmp/test_recursive/level1/level2/level3",
			wantErr: false,
		},
		{
			name:    "create single directory",
			path:    "/tmp/test_recursive_single",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			baseDir := "/tmp/test_recursive"
			if strings.Contains(tt.path, "test_recursive_single") {
				baseDir = "/tmp/test_recursive_single"
			}
			os.RemoveAll(baseDir)
			defer os.RemoveAll("/tmp/test_recursive")
			defer os.RemoveAll("/tmp/test_recursive_single")

			err := createDirRecursive(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDirRecursive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify directory exists
			if _, err := os.Stat(tt.path); os.IsNotExist(err) && !tt.wantErr {
				t.Errorf("Directory was not created: %v", tt.path)
			}
		})
	}
}

// TestDeleteDir tests the deleteDir function.
func TestDeleteDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "delete existing directory",
			path:    "/tmp/test_delete_dir",
			wantErr: false,
		},
		{
			name:    "delete non-existing directory",
			path:    "/tmp/test_delete_dir_nonexist",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: create directory for existing test
			if strings.Contains(tt.name, "existing") && !strings.Contains(tt.name, "non-existing") {
				if err := os.MkdirAll(tt.path, os.ModePerm); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			err := deleteDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify directory doesn't exist
			if _, err := os.Stat(tt.path); !os.IsNotExist(err) && !tt.wantErr {
				t.Errorf("Directory was not deleted: %v", tt.path)
			}
		})
	}
}

// TestCopyFile tests the copyFile function.
func TestCopyFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		setupErr  bool
		wantErr   bool
		errString string
	}{
		{
			name:    "copy regular file",
			content: "test content\nline 2\nline 3",
			wantErr: false,
		},
		{
			name:      "copy non-existing file",
			content:   "",
			setupErr:  true,
			wantErr:   true,
			errString: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcPath := "/tmp/test_copy_src.txt"
			dstPath := "/tmp/test_copy_dst.txt"
			defer os.Remove(srcPath)
			defer os.Remove(dstPath)

			// Setup: create source file
			if !tt.setupErr {
				err := os.WriteFile(srcPath, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			err := copyFile(srcPath, dstPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errString != "" {
				if !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("Expected error containing %q, got %q", tt.errString, err.Error())
				}
				return
			}

			if !tt.wantErr {
				// Verify file was copied correctly
				content, err := os.ReadFile(dstPath)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
					return
				}
				if string(content) != tt.content {
					t.Errorf("File content mismatch. Got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

// TestReadFile tests the readFile function.
func TestReadFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
		wantErr bool
	}{
		{
			name:    "read multi-line file",
			content: "line 1\nline 2\nline 3",
			want:    []string{"line 1", "line 2", "line 3"},
			wantErr: false,
		},
		{
			name:    "read empty file",
			content: "",
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "read single line file",
			content: "single line",
			want:    []string{"single line"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := "/tmp/test_read_file.txt"
			defer os.Remove(filePath)

			// Setup: create file with content
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			got, err := readFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("readFile() returned %d lines, want %d lines", len(got), len(tt.want))
				return
			}

			for i, line := range got {
				if line != tt.want[i] {
					t.Errorf("readFile() line %d = %q, want %q", i, line, tt.want[i])
				}
			}
		})
	}
}

// TestReadFileError tests the readFile function with non-existing file.
func TestReadFileError(t *testing.T) {
	_, err := readFile("/tmp/non_existing_file_12345.txt")
	if err == nil {
		t.Error("readFile() expected error for non-existing file, got nil")
	}
}

// TestWriteFile tests the writeFile function.
func TestWriteFile(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name:    "write text content",
			content: []byte("test content"),
			wantErr: false,
		},
		{
			name:    "write empty content",
			content: []byte(""),
			wantErr: false,
		},
		{
			name:    "write JSON content",
			content: []byte(`{"key": "value"}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := "/tmp/test_write_file.txt"
			defer os.Remove(filePath)

			err := writeFile(filePath, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify file was written correctly
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to read written file: %v", err)
				return
			}
			if string(content) != string(tt.content) {
				t.Errorf("File content mismatch. Got %q, want %q", string(content), string(tt.content))
			}
		})
	}
}

// TestGetShaSumContents tests the getShaSumContents function.
func TestGetShaSumContents(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    [][]string
		wantErr bool
	}{
		{
			name: "parse valid SHA256SUMS",
			content: "abc123  terraform-provider-example_1.0.0_linux_amd64.zip\n" +
				"def456  terraform-provider-example_1.0.0_darwin_amd64.zip\n" +
				"ghi789  terraform-provider-example_1.0.0_windows_amd64.zip",
			want: [][]string{
				{"abc123", "terraform-provider-example_1.0.0_linux_amd64.zip"},
				{"def456", "terraform-provider-example_1.0.0_darwin_amd64.zip"},
				{"ghi789", "terraform-provider-example_1.0.0_windows_amd64.zip"},
			},
			wantErr: false,
		},
		{
			name:    "parse single entry",
			content: "abc123  terraform-provider-example_1.0.0_linux_amd64.zip",
			want: [][]string{
				{"abc123", "terraform-provider-example_1.0.0_linux_amd64.zip"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory and file
			distPath := "/tmp/test_dist"
			repoName := "terraform-provider-example"
			version := "1.0.0"
			shaSumFileName := repoName + "_" + version + "_SHA256SUMS"

			if err := os.MkdirAll(distPath, os.ModePerm); err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}
			defer os.RemoveAll(distPath)

			shaSumPath := filepath.Join(distPath, shaSumFileName)
			err := os.WriteFile(shaSumPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to setup test: %v", err)
			}

			got, err := getShaSumContents(distPath, repoName, version)
			if (err != nil) != tt.wantErr {
				t.Errorf("getShaSumContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("getShaSumContents() returned %d entries, want %d entries", len(got), len(tt.want))
				return
			}

			for i, entry := range got {
				if len(entry) != 2 {
					t.Errorf("Entry %d has %d elements, want 2", i, len(entry))
					continue
				}
				if entry[0] != tt.want[i][0] || entry[1] != tt.want[i][1] {
					t.Errorf("Entry %d = %v, want %v", i, entry, tt.want[i])
				}
			}
		})
	}
}

// TestGetShaSumContentsError tests error case for getShaSumContents.
func TestGetShaSumContentsError(t *testing.T) {
	_, err := getShaSumContents("/tmp/nonexistent", "repo", "1.0.0")
	if err == nil {
		t.Error("getShaSumContents() expected error for non-existing file, got nil")
	}
}

// TestPlatformJSON tests Platform struct JSON marshalling.
func TestPlatformJSON(t *testing.T) {
	platform := Platform{
		Os:   "linux",
		Arch: "amd64",
	}

	data, err := json.Marshal(platform)
	if err != nil {
		t.Fatalf("Failed to marshal Platform: %v", err)
	}

	var decoded Platform
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Platform: %v", err)
	}

	if decoded.Os != platform.Os || decoded.Arch != platform.Arch {
		t.Errorf("Platform mismatch after marshal/unmarshal. Got %+v, want %+v", decoded, platform)
	}
}

// TestVersionJSON tests Version struct JSON marshalling.
func TestVersionJSON(t *testing.T) {
	version := Version{
		Version:   "1.0.0",
		Protocols: []string{"4.0", "5.1"},
		Platforms: []Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
	}

	data, err := json.Marshal(version)
	if err != nil {
		t.Fatalf("Failed to marshal Version: %v", err)
	}

	var decoded Version
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Version: %v", err)
	}

	if decoded.Version != version.Version {
		t.Errorf("Version mismatch. Got %v, want %v", decoded.Version, version.Version)
	}
	if len(decoded.Protocols) != len(version.Protocols) {
		t.Errorf("Protocols length mismatch. Got %d, want %d", len(decoded.Protocols), len(version.Protocols))
	}
	if len(decoded.Platforms) != len(version.Platforms) {
		t.Errorf("Platforms length mismatch. Got %d, want %d", len(decoded.Platforms), len(version.Platforms))
	}
}

// TestVersionsJSON tests Versions struct JSON marshalling.
func TestVersionsJSON(t *testing.T) {
	versions := Versions{
		Versions: []Version{
			{
				Version:   "1.0.0",
				Protocols: []string{"4.0", "5.1"},
				Platforms: []Platform{{Os: "linux", Arch: "amd64"}},
			},
			{
				Version:   "1.1.0",
				Protocols: []string{"5.0", "5.1"},
				Platforms: []Platform{{Os: "darwin", Arch: "arm64"}},
			},
		},
	}

	data, err := json.Marshal(versions)
	if err != nil {
		t.Fatalf("Failed to marshal Versions: %v", err)
	}

	var decoded Versions
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Versions: %v", err)
	}

	if len(decoded.Versions) != len(versions.Versions) {
		t.Errorf("Versions length mismatch. Got %d, want %d", len(decoded.Versions), len(versions.Versions))
	}
}

// TestWellKnownJSON tests WellKnown struct JSON marshalling.
func TestWellKnownJSON(t *testing.T) {
	wellKnown := WellKnown{
		ProvidersV1: "/v1/providers/",
		ModulesV1:   "/v1/modules/",
	}

	data, err := json.Marshal(wellKnown)
	if err != nil {
		t.Fatalf("Failed to marshal WellKnown: %v", err)
	}

	var decoded WellKnown
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal WellKnown: %v", err)
	}

	if decoded.ProvidersV1 != wellKnown.ProvidersV1 || decoded.ModulesV1 != wellKnown.ModulesV1 {
		t.Errorf("WellKnown mismatch. Got %+v, want %+v", decoded, wellKnown)
	}
}

// TestArchitectureJSON tests Architecture struct JSON marshalling.
func TestArchitectureJSON(t *testing.T) {
	architecture := Architecture{
		Protocols:           []string{"4.0", "5.1"},
		Os:                  "linux",
		Arch:                "amd64",
		Filename:            "terraform-provider-example_1.0.0_linux_amd64.zip",
		DownloadUrl:         "https://example.com/download/terraform-provider-example_1.0.0_linux_amd64.zip",
		ShasumsUrl:          "https://example.com/SHA256SUMS",
		ShasumsSignatureUrl: "https://example.com/SHA256SUMS.sig",
		Shasum:              "abc123def456",
	}

	data, err := json.Marshal(architecture)
	if err != nil {
		t.Fatalf("Failed to marshal Architecture: %v", err)
	}

	var decoded Architecture
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal Architecture: %v", err)
	}

	if decoded.Os != architecture.Os {
		t.Errorf("Os mismatch. Got %v, want %v", decoded.Os, architecture.Os)
	}
	if decoded.Arch != architecture.Arch {
		t.Errorf("Arch mismatch. Got %v, want %v", decoded.Arch, architecture.Arch)
	}
	if decoded.Filename != architecture.Filename {
		t.Errorf("Filename mismatch. Got %v, want %v", decoded.Filename, architecture.Filename)
	}
	if decoded.Shasum != architecture.Shasum {
		t.Errorf("Shasum mismatch. Got %v, want %v", decoded.Shasum, architecture.Shasum)
	}
}

// TestDefaultWellKnownData tests the default well-known data.
func TestDefaultWellKnownData(t *testing.T) {
	if defaultWellKnownData.ProvidersV1 != "/v1/providers/" {
		t.Errorf("ProvidersV1 = %v, want /v1/providers/", defaultWellKnownData.ProvidersV1)
	}
	if defaultWellKnownData.ModulesV1 != "/v1/modules/" {
		t.Errorf("ModulesV1 = %v, want /v1/modules/", defaultWellKnownData.ModulesV1)
	}
}
