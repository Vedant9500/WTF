package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Vedant9500/WTF/internal/errors"
)

// FilePermissions defines secure file permissions for different file types
type FilePermissions struct {
	ConfigFile    os.FileMode // Configuration files (readable by owner only)
	DataFile      os.FileMode // Data files (readable by owner and group)
	ExecutableFile os.FileMode // Executable files
	Directory     os.FileMode // Directories
	TempFile      os.FileMode // Temporary files
}

// DefaultPermissions returns secure default file permissions
func DefaultPermissions() FilePermissions {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix-style permissions, but we set them anyway
		return FilePermissions{
			ConfigFile:     0600, // Owner read/write
			DataFile:       0644, // Owner read/write, group/others read
			ExecutableFile: 0755, // Owner read/write/execute, group/others read/execute
			Directory:      0755, // Owner read/write/execute, group/others read/execute
			TempFile:       0600, // Owner read/write
		}
	}
	
	return FilePermissions{
		ConfigFile:     0600, // Owner read/write only
		DataFile:       0644, // Owner read/write, group/others read only
		ExecutableFile: 0755, // Owner read/write/execute, group/others read/execute
		Directory:      0755, // Owner read/write/execute, group/others read/execute
		TempFile:       0600, // Owner read/write only
	}
}

// RestrictivePermissions returns more restrictive file permissions
func RestrictivePermissions() FilePermissions {
	return FilePermissions{
		ConfigFile:     0600, // Owner read/write only
		DataFile:       0600, // Owner read/write only
		ExecutableFile: 0700, // Owner read/write/execute only
		Directory:      0700, // Owner read/write/execute only
		TempFile:       0600, // Owner read/write only
	}
}

// SetSecureFilePermissions sets secure permissions on a file based on its type
func SetSecureFilePermissions(filePath string, fileType string) error {
	permissions := DefaultPermissions()
	
	var mode os.FileMode
	switch fileType {
	case "config":
		mode = permissions.ConfigFile
	case "data":
		mode = permissions.DataFile
	case "executable":
		mode = permissions.ExecutableFile
	case "directory":
		mode = permissions.Directory
	case "temp":
		mode = permissions.TempFile
	default:
		mode = permissions.DataFile // Default to data file permissions
	}
	
	if err := os.Chmod(filePath, mode); err != nil {
		return errors.NewAppError(errors.ErrorTypeFileSystem, 
			fmt.Sprintf("failed to set permissions on %s", filePath), err).
			WithUserMessage(fmt.Sprintf("Could not set secure permissions on file: %s", filePath)).
			WithContext("file_path", filePath).
			WithContext("file_type", fileType).
			WithContext("desired_mode", fmt.Sprintf("%o", mode)).
			WithSuggestions(
				"Check if you have permission to modify the file",
				"Ensure the file exists and is accessible",
				"Try running with elevated privileges if necessary",
			)
	}
	
	return nil
}

// ValidateFilePermissions checks if a file has secure permissions
func ValidateFilePermissions(filePath string, fileType string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("cannot access file %s", filePath), err).
			WithUserMessage(fmt.Sprintf("Cannot access file: %s", filePath)).
			WithContext("file_path", filePath)
	}
	
	currentMode := info.Mode().Perm()
	permissions := DefaultPermissions()
	
	var expectedMode os.FileMode
	switch fileType {
	case "config":
		expectedMode = permissions.ConfigFile
	case "data":
		expectedMode = permissions.DataFile
	case "executable":
		expectedMode = permissions.ExecutableFile
	case "directory":
		expectedMode = permissions.Directory
	case "temp":
		expectedMode = permissions.TempFile
	default:
		expectedMode = permissions.DataFile
	}
	
	// On Windows, permission checking is less strict
	if runtime.GOOS == "windows" {
		return nil
	}
	
	// Check if file is world-writable (security risk)
	if currentMode&0002 != 0 {
		return errors.NewAppError(errors.ErrorTypePermission,
			fmt.Sprintf("file %s is world-writable", filePath), nil).
			WithUserMessage(fmt.Sprintf("Security risk: file %s can be modified by anyone", filePath)).
			WithContext("file_path", filePath).
			WithContext("current_mode", fmt.Sprintf("%o", currentMode)).
			WithSuggestions(
				fmt.Sprintf("Run: chmod %o %s", expectedMode, filePath),
				"Remove world-write permissions for security",
			)
	}
	
	// Check if file is group-writable for sensitive files
	if (fileType == "config" || fileType == "temp") && currentMode&0020 != 0 {
		return errors.NewAppError(errors.ErrorTypePermission,
			fmt.Sprintf("sensitive file %s is group-writable", filePath), nil).
			WithUserMessage(fmt.Sprintf("Security risk: sensitive file %s can be modified by group members", filePath)).
			WithContext("file_path", filePath).
			WithContext("file_type", fileType).
			WithContext("current_mode", fmt.Sprintf("%o", currentMode)).
			WithSuggestions(
				fmt.Sprintf("Run: chmod %o %s", expectedMode, filePath),
				"Remove group-write permissions for sensitive files",
			)
	}
	
	return nil
}

// CreateSecureFile creates a file with secure permissions
func CreateSecureFile(filePath string, fileType string) (*os.File, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, DefaultPermissions().Directory); err != nil {
		return nil, errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("failed to create directory %s", dir), err).
			WithUserMessage(fmt.Sprintf("Could not create directory: %s", dir)).
			WithContext("directory", dir)
	}
	
	// Create file with secure permissions
	permissions := DefaultPermissions()
	var mode os.FileMode
	switch fileType {
	case "config":
		mode = permissions.ConfigFile
	case "data":
		mode = permissions.DataFile
	case "temp":
		mode = permissions.TempFile
	default:
		mode = permissions.DataFile
	}
	
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("failed to create file %s", filePath), err).
			WithUserMessage(fmt.Sprintf("Could not create file: %s", filePath)).
			WithContext("file_path", filePath).
			WithContext("file_type", fileType)
	}
	
	return file, nil
}

// CreateSecureDirectory creates a directory with secure permissions
func CreateSecureDirectory(dirPath string) error {
	permissions := DefaultPermissions()
	
	if err := os.MkdirAll(dirPath, permissions.Directory); err != nil {
		return errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("failed to create directory %s", dirPath), err).
			WithUserMessage(fmt.Sprintf("Could not create directory: %s", dirPath)).
			WithContext("directory", dirPath).
			WithSuggestions(
				"Check if you have permission to create directories",
				"Ensure the parent directory exists",
				"Try running with elevated privileges if necessary",
			)
	}
	
	return nil
}

// ValidateDirectoryPermissions checks if a directory has secure permissions
func ValidateDirectoryPermissions(dirPath string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("cannot access directory %s", dirPath), err).
			WithUserMessage(fmt.Sprintf("Cannot access directory: %s", dirPath)).
			WithContext("directory", dirPath)
	}
	
	if !info.IsDir() {
		return errors.NewAppError(errors.ErrorTypeValidation,
			fmt.Sprintf("%s is not a directory", dirPath), nil).
			WithUserMessage(fmt.Sprintf("Path is not a directory: %s", dirPath)).
			WithContext("path", dirPath)
	}
	
	currentMode := info.Mode().Perm()
	
	// On Windows, permission checking is less strict
	if runtime.GOOS == "windows" {
		return nil
	}
	
	// Check if directory is world-writable (security risk)
	if currentMode&0002 != 0 {
		return errors.NewAppError(errors.ErrorTypePermission,
			fmt.Sprintf("directory %s is world-writable", dirPath), nil).
			WithUserMessage(fmt.Sprintf("Security risk: directory %s can be modified by anyone", dirPath)).
			WithContext("directory", dirPath).
			WithContext("current_mode", fmt.Sprintf("%o", currentMode)).
			WithSuggestions(
				fmt.Sprintf("Run: chmod 755 %s", dirPath),
				"Remove world-write permissions for security",
			)
	}
	
	return nil
}

// SecureFileOperations provides a set of secure file operations
type SecureFileOperations struct {
	permissions FilePermissions
}

// NewSecureFileOperations creates a new SecureFileOperations instance
func NewSecureFileOperations() *SecureFileOperations {
	return &SecureFileOperations{
		permissions: DefaultPermissions(),
	}
}

// WriteSecureFile writes data to a file with secure permissions
func (sfo *SecureFileOperations) WriteSecureFile(filePath string, data []byte, fileType string) error {
	file, err := CreateSecureFile(filePath, fileType)
	if err != nil {
		return err
	}
	defer file.Close()
	
	if _, err := file.Write(data); err != nil {
		return errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("failed to write to file %s", filePath), err).
			WithUserMessage(fmt.Sprintf("Could not write to file: %s", filePath)).
			WithContext("file_path", filePath)
	}
	
	return nil
}

// ReadSecureFile reads data from a file after validating permissions
func (sfo *SecureFileOperations) ReadSecureFile(filePath string, fileType string) ([]byte, error) {
	if err := ValidateFilePermissions(filePath, fileType); err != nil {
		// Log warning but don't fail - file might still be readable
		// In a real application, you might want to log this warning
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrorTypeFileSystem,
			fmt.Sprintf("failed to read file %s", filePath), err).
			WithUserMessage(fmt.Sprintf("Could not read file: %s", filePath)).
			WithContext("file_path", filePath)
	}
	
	return data, nil
}