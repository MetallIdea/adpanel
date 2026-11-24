package filesystem

import "errors"

var (
	// ErrPathTraversal возвращётся при попытке выйти за пределы RootDir
	// через ../ или абсолютный путь.
	ErrPathTraversal = errors.New("access denied: path traversal detected")

	// ErrSymlinkTraversal возвращётся при обнаружении символической ссылки,
	// ведущей за пределы RootDir.
	ErrSymlinkTraversal = errors.New("access denied: symlink traversal detected")

	// ErrFileTooLarge возвращётся при загрузке файла, превышающего MaxUploadSize.
	ErrFileTooLarge = errors.New("file too large")

	// ErrDirectoryNotFound возвращётся, если запрошенная директория не существует.
	ErrDirectoryNotFound = errors.New("directory not found")

	// ErrFileAlreadyExists возвращётся при попытке загрузить файл с уже
	// существующим именем (перезапись запрещена).
	ErrFileAlreadyExists = errors.New("file already exists")

	// ErrInvalidMultipartRequest возвращётся при некорректном multipart-запросе.
	ErrInvalidMultipartRequest = errors.New("invalid multipart request")
)
