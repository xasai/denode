package shared

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ricochet2200/go-disk-usage/du"
)

var (
	WorkDirPath    string
	AccsDirPath    string
	NodeAddr       []byte
	WorkDirName    = "dfile"
	ConfDirName    = "config"
	StorageDirName = "storage"
)

func GetAvailableSpace(storagePath string) int {
	var KB = uint64(1024)
	usage := du.NewDiskUsage(storagePath)
	return int(usage.Free() / (KB * KB * KB))
}

// ====================================================================================

func InitPaths() error {
	const info = "shared.InitPaths"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%s %w", info, err)
	}

	WorkDirPath = filepath.Join(homeDir, WorkDirName)

	AccsDirPath = filepath.Join(WorkDirPath, "accounts")

	return nil

}

// ====================================================================================

func CreateIfNotExistAccDirs() error {
	const info = "shared.CreateIfNotExistAccDirs"
	_, err := os.Stat(WorkDirPath)
	if err != nil {
		errPart := strings.Split(err.Error(), ":")

		if strings.Trim(errPart[1], " ") != "no such file or directory" {
			return fmt.Errorf("%s %w", info, err)
		}

		err = os.MkdirAll(WorkDirPath, os.ModePerm|os.ModeDir)
		if err != nil {
			return fmt.Errorf("%s %w", info, err)
		}
	}

	_, err = os.Stat(AccsDirPath)
	if err != nil {
		errPart := strings.Split(err.Error(), ":")

		if strings.Trim(errPart[1], " ") != "no such file or directory" {
			return fmt.Errorf("%s %w", info, err)
		}

		err = os.MkdirAll(AccsDirPath, os.ModePerm|os.ModeDir)
		if err != nil {
			return fmt.Errorf("%s %w", info, err)

		}
	}

	return nil
}

// ====================================================================================

func ContainsAccount(accounts []string, address string) bool {
	for _, a := range accounts {
		if a == address {
			return true
		}
	}
	return false
}

// ====================================================================================

func ReadFromConsole() (string, error) {
	const info = "shared.ReadFromConsole"
	fmt.Print("Enter value here: ")
	reader := bufio.NewReader(os.Stdin)
	// ReadString will block until the delimiter is entered
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("%s %w", info, err)
	}

	// remove the delimiter from the string
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")

	return input, nil
}

// ====================================================================================

func CalcRootHash(hashArr []string) (string, [][][]byte, error) {
	const info = "shared.CalcRootHash"
	resByte := [][][]byte{}
	base := [][]byte{}

	emptyValue, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		return "", resByte, fmt.Errorf("%s %w", info, err)
	}

	for _, v := range hashArr {
		decoded, err := hex.DecodeString(v)
		if err != nil {
			return "", resByte, fmt.Errorf("%s %w", info, err)
		}
		base = append(base, decoded)
	}

	if len(base)%2 != 0 {
		base = append(base, emptyValue)
	}

	resByte = append(resByte, base)

	for len(resByte[len(resByte)-1]) != 1 {
		prevList := resByte[len(resByte)-1]
		resByte = append(resByte, [][]byte{})
		r := len(prevList) / 2

		for i := 0; i < r; i++ {
			a := prevList[i*2]
			b := prevList[i*2+1]

			concatBytes := append(a, b...)
			hSum := sha256.Sum256(concatBytes)

			resByte[len(resByte)-1] = append(resByte[len(resByte)-1], hSum[:])

		}

		if len(resByte[len(resByte)-1])%2 != 0 && len(prevList) > 2 {
			resByte[len(resByte)-1] = append(resByte[len(resByte)-1], emptyValue)

		}
	}

	return hex.EncodeToString(resByte[len(resByte)-1][0]), resByte, nil
}

// ====================================================================================

func encryptAES(key, data []byte) ([]byte, error) {
	const info = "shared.encryptAES"
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s %w", info, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%s %w", info, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("%s %w", info, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil

}

// ====================================================================================

func decryptAES(key, data []byte) ([]byte, error) {
	const info = "shared.decryptAES"
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s %w", info, err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%s %w", info, err)
	}
	nonce, encrData := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	decrData, err := gcm.Open(nil, nonce, encrData, nil)
	if err != nil {
		return nil, fmt.Errorf("%s %w", info, err)
	}

	return decrData, nil
}

// ====================================================================================

func GetDeviceMacAddr() (string, error) {
	const info = "shared.GetDeviceMacAddr"
	var addr string
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("%s %w", info, err)
	}

	for _, i := range interfaces {
		if !bytes.Equal(i.HardwareAddr, nil) {
			addr = i.HardwareAddr.String()
			break
		}
	}

	return addr, nil
}

// ====================================================================================

func EncryptNodeAddr(addr common.Address) ([]byte, error) {
	const info = "shared.EncryptNodeAddr"
	var nodeAddr []byte

	macAddr, err := GetDeviceMacAddr()
	if err != nil {
		return nodeAddr, fmt.Errorf("%s %w", info, err)
	}

	encrKey := sha256.Sum256([]byte(macAddr))

	encryptedAddr, err := encryptAES(encrKey[:], addr.Bytes())
	if err != nil {
		return nodeAddr, fmt.Errorf("%s %w", info, err)
	}

	return encryptedAddr, nil
}

// ====================================================================================

func DecryptNodeAddr() (common.Address, error) {
	const info = "shared.DecryptNodeAddr"
	var nodeAddr common.Address

	macAddr, err := GetDeviceMacAddr()
	if err != nil {
		return nodeAddr, fmt.Errorf("%s %w", info, err)
	}

	encrKey := sha256.Sum256([]byte(macAddr))

	accAddr, err := decryptAES(encrKey[:], NodeAddr)
	if err != nil {
		return nodeAddr, fmt.Errorf("%s %w", info, err)
	}

	return common.BytesToAddress(accAddr), nil
}

// ====================================================================================

func LogError(errMsg string) error {
	logsFile, err := os.OpenFile("./errorLogs", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0700)
	if err != nil {
		return err
	}

	defer logsFile.Close()

	currentTime := time.Now().Local()

	logMsg := currentTime.String() + ": " + errMsg + "\n"

	fmt.Println(logMsg) //TODO remove

	_, err = logsFile.WriteString(logMsg)
	if err != nil {
		return err
	}

	return nil
}

// ====================================================================================

// ====================================================================================
