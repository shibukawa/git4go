package git4go

type ErrorCode int

const (
	// Requested object could not be found
	ErrNotFound ErrorCode = -3
	// Operation not allowed on bare repository
	ErrBareRepository ErrorCode = -8
	// The operation is not valid for a directory
	ErrDirectory ErrorCode = -23
	// Signals end of iteration with iterator
	ErrIterOver ErrorCode = -31
)

type GitError struct {
	Message string
	Code    ErrorCode
}

func (e GitError) Error() string {
	return e.Message
}

func IsErrorCode(err error, c ErrorCode) bool {
	if err == nil {
		return false
	}
	if gitError, ok := err.(*GitError); ok {
		return gitError.Code == c
	}
	return false
}

func MakeGitError(message string, errorCode ErrorCode) error {
	return &GitError{
		Message: message,
		Code:    errorCode,
	}
}

const (
	GitOidRawSize                    = 20
	GitOidHexSize                    = 40
	GitOidMinimumPrefixLength        = 4
	GitObjectDirMode          uint32 = 0777
	GitObjectFileMode         uint32 = 0444
)

func Discover(start string, acrossFs bool, ceilingDirs []string) (string, error) {
	var flags uint32 = 0
	if acrossFs {
		flags = GIT_REPOSITORY_OPEN_CROSS_FS
	}
	repoPath, _, _, err := findRepo(start, flags, ceilingDirs)
	return repoPath, err
}
