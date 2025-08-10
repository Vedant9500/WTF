package validation

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPermissions(t *testing.T) {
	perms := DefaultPermissions()
	
	// Test that permissions are set
	if perms.ConfigFile == 0 {
		t.Error("ConfigFile permissions should not be zero")
	}
	if perms.DataFile == 0 {
		t.Error("DataFile permissions should not be zero")
	}
	if perms.ExecutableFile == 0 {
		t.Error("ExecutableFile permissions should not be zero")
	}
	if perms.Directory == 0 {
		t.Error("Directory permissions should not be zero")
	}
	if perms.TempFile == 0 {
		t.Error("TempFile permissions should not be zero")
	}
	
	// Test that config files are more restrictive than data files
	if perms.ConfigFile > perms.DataFile {
		t.Error("Config files should have more restrictive permissions than data files")
	}
}

func TestRestrictivePermissions(t *testing.T) {
	restrictive := RestrictivePermissions()
	default_ := DefaultPermissions()
	
	// Restrictive permissions should be more restrictive or equal
	if restrictive.ConfigFile > default_.ConfigFile {
		t.Error("Restrictive config permissions should be more restrictive")
	}
	if restrictive.DataFile > default_.DataFile {
		t.Error("Restrictive data permissions should be more restrictive")
	}
	if restrictive.ExecutableFile > default_.ExecutableFile {
		t.Error("Restrictive executable permissions should be more restrictive")
	}
}

func TestCreateSecureFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "wtf_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testFile := filepath.Join(tempDir, "test_config.yml")
	
	// Create secure file
	file, err := CreateSecureFile(testFile, "config")
	if err != nil {
		t.Fatalf("Failed to create secure file: %v", err)
	}
	file.Close()
	
	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Secure file was not created")
	}
	
	// On Unix systems, verify permissions
	if runtime.GOOS != "windows" {
		info, err := os.Stat(testFile)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		
		expectedMode := DefaultPermissions().ConfigFile
		actualMode := info.Mode().Perm()
		
		if actualMode != expectedMode {
			t.Errorf("Expected file mode %o, got %o", expectedMode, actualMode)
		}
	}
}

func TestCreateSecureDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "wtf_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testDir := filepath.Join(tempDir, "secure_dir")
	
	// Create secure directory
	err = CreateSecureDirectory(testDir)
	if err != nil {
		t.Fatalf("Failed to create secure directory: %v", err)
	}
	
	// Verify directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	
	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
	
	// On Unix systems, verify permissions
	if runtime.GOOS != "windows" {
		expectedMode := DefaultPermissions().Directory
		actualMode := info.Mode().Perm()
		
		if actualMode != expectedMode {
			t.Errorf("Expected directory mode %o, got %o", expectedMode, actualMode)
		}
	}
}

func TestSetSecureFilePermissions(t *testing.T) {
	// Skip on Windows as it doesn't support Unix-style permissions
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "wtf_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// Set insecure permissions first
	err = os.Chmod(tempFile.Name(), 0777)
	if err != nil {
		t.Fatalf("Failed to set initial permissions: %v", err)
	}
	
	// Set secure permissions
	err = SetSecureFilePermissions(tempFile.Name(), "config")
	if err != nil {
		t.Fatalf("Failed to set secure permissions: %v", err)
	}
	
	// Verify permissions were set correctly
	info, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	
	expectedMode := DefaultPermissions().ConfigFile
	actualMode := info.Mode().Perm()
	
	if actualMode != expectedMode {
		t.Errorf("Expected file mode %o, got %o", expectedMode, actualMode)
	}
}

func TestValidateFilePermissions(t *testing.T) {
	// Skip on Windows as it doesn't support Unix-style permissions
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission validation test on Windows")
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "wtf_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// Test with secure permissions
	err = os.Chmod(tempFile.Name(), 0600)
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}
	
	err = ValidateFilePermissions(tempFile.Name(), "config")
	if err != nil {
		t.Errorf("Validation failed for secure permissions: %v", err)
	}
	
	// Test with world-writable permissions (should fail)
	err = os.Chmod(tempFile.Name(), 0666)
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}
	
	err = ValidateFilePermissions(tempFile.Name(), "config")
	if err == nil {
		t.Error("Validation should have failed for world-writable file")
	}
	
	// Test with group-writable permissions for config file (should fail)
	err = os.Chmod(tempFile.Name(), 0620)
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}
	
	err = ValidateFilePermissions(tempFile.Name(), "config")
	if err == nil {
		t.Error("Validation should have failed for group-writable config file")
	}
}

func TestValidateDirectoryPermissions(t *testing.T) {
	// Skip on Windows as it doesn't support Unix-style permissions
	if runtime.GOOS == "windows" {
		t.Skip("Skipping directory permission validation test on Windows")
	}
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "wtf_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test with secure permissions
	err = os.Chmod(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}
	
	err = ValidateDirectoryPermissions(tempDir)
	if err != nil {
		t.Errorf("Validation failed for secure directory permissions: %v", err)
	}
	
	// Test with world-writable permissions (should fail)
	err = os.Chmod(tempDir, 0777)
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}
	
	err = ValidateDirectoryPermissions(tempDir)
	if err == nil {
		t.Error("Validation should have failed for world-writable directory")
	}
}

func TestSecureFileOperations(t *testing.T) {
	sfo := NewSecureFileOperations()
	
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "wtf_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testFile := filepath.Join(tempDir, "test_data.txt")
	testData := []byte("test data content")
	
	// Write secure file
	err = sfo.WriteSecureFile(testFile, testData, "data")
	if err != nil {
		t.Fatalf("Failed to write secure file: %v", err)
	}
	
	// Read secure file
	readData, err := sfo.ReadSecureFile(testFile, "data")
	if err != nil {
		t.Fatalf("Failed to read secure file: %v", err)
	}
	
	// Verify data matches
	if string(readData) != string(testData) {
		t.Errorf("Expected data '%s', got '%s'", string(testData), string(readData))
	}
	
	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Secure file was not created")
	}
}

func TestSecureFileOperationsWithNonExistentFile(t *testing.T) {
	sfo := NewSecureFileOperations()
	
	// Try to read non-existent file
	_, err := sfo.ReadSecureFile("/non/existent/file.txt", "data")
	if err == nil {
		t.Error("Reading non-existent file should have failed")
	}
}

func TestPermissionFileTypes(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "wtf_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	fileTypes := []string{"config", "data", "executable", "temp", "unknown"}
	
	for _, fileType := range fileTypes {
		t.Run(fileType, func(t *testing.T) {
			testFile := filepath.Join(tempDir, "test_"+fileType+".txt")
			
			// Create file with specific type
			file, err := CreateSecureFile(testFile, fileType)
			if err != nil {
				t.Fatalf("Failed to create secure file for type %s: %v", fileType, err)
			}
			file.Close()
			
			// Verify file was created
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Errorf("File was not created for type %s", fileType)
			}
			
			// Set permissions based on type
			err = SetSecureFilePermissions(testFile, fileType)
			if err != nil {
				t.Errorf("Failed to set permissions for type %s: %v", fileType, err)
			}
		})
	}
}

func TestPermissionValidationEdgeCases(t *testing.T) {
	// Test validation of non-existent file
	err := ValidateFilePermissions("/non/existent/file.txt", "data")
	if err == nil {
		t.Error("Validation of non-existent file should have failed")
	}
	
	// Test validation of non-existent directory
	err = ValidateDirectoryPermissions("/non/existent/directory")
	if err == nil {
		t.Error("Validation of non-existent directory should have failed")
	}
	
	// Create temporary file and test validation as directory
	tempFile, err := os.CreateTemp("", "wtf_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	err = ValidateDirectoryPermissions(tempFile.Name())
	if err == nil {
		t.Error("Validation of file as directory should have failed")
	}
}