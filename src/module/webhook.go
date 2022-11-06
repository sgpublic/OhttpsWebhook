package module

import (
	"OhttpsWebhook/src/config"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
)

func DoWebhook(path string, port int) {
	http.HandleFunc(path, _HandleWebhook)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatalf("Service start failed: %v", err)
	}
}

var resp = func() []byte {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	_ = jsonEncoder.Encode(struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
	return bf.Bytes()
}()

func _HandleWebhook(w http.ResponseWriter, r *http.Request) {
	data := new(Ohttps)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(data)
	if err != nil {
		log.Warningf("Read request from Ohttps failed: %v", err)
	} else {
		_Process(*data)
	}
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("response failed: %v", err)
	}
}

type Ohttps struct {
	Timestamp int `json:"timestamp"`
	Payload   struct {
		CertificateDomain        string `json:"certificateDomains"`
		CertificateCertKey       string `json:"certificateCertKey"`
		CertificateFullchainCert string `json:"certificateFullchainCerts"`
	} `json:"payload"`
	Sign string `json:"sign"`
}

func _Process(data Ohttps) {
	log.Infof("Processing request, domain: " + data.Payload.CertificateDomain + ", sign: " + data.Sign)
	target, err := config.GetTarget(data.Payload.CertificateDomain)
	if err != nil {
		log.Warningf("Error occupied during the process: %v", err)
		return
	}

	// 校验 sign
	h := md5.New()
	h.Write([]byte(strconv.Itoa(data.Timestamp) + ":" + config.GetWebhookKey()))
	sign := hex.EncodeToString(h.Sum(nil))
	if data.Sign != sign {
		log.Warningf("Sign mismatch! domain: %s", data.Payload.CertificateDomain)
		return
	}

	CertKey, err := os.OpenFile(target.CertKey, os.O_RDWR, 0311)
	if err != nil && os.IsNotExist(err) {
		log.Warningf("Target CertKey not found, create new one...")
		CertKey, err = os.Create(target.CertKey)
		if err != nil {
			log.Warningf("CertKey create failed: %v", err)
			return
		}
	} else {
		log.Infof("Backing up old CertKey...")
		backup, err := os.OpenFile(target.CertKey+".bak", os.O_WRONLY|os.O_CREATE, 0311)
		if err != nil {
			log.Warningf("CertKey backup create failed: %v", err)
			return
		}
		_, err = io.Copy(backup, CertKey)
		if err != nil {
			log.Warningf("CertKey backup write failed: %v", err)
			return
		}
	}

	rollback := func() {
		backup, err2 := os.OpenFile(target.CertKey+".bak", os.O_RDONLY, 0311)
		if err2 != nil && os.IsNotExist(err2) {
			log.Warningf("CertKey backup up not exists, skip rollback.")
			return
		}
		CertKey, err2 := os.OpenFile(target.CertKey+".bak", os.O_WRONLY|os.O_CREATE, 0311)
		if err2 != nil {
			log.Warningf("CertKey create failed: %v", err2)
			return
		}
		_, err2 = io.Copy(CertKey, backup)
		if err2 != nil {
			log.Warningf("CertKey rollback failed: %v", err)
		}
	}

	FullchainCerts, err := os.OpenFile(target.FullchainCerts, os.O_RDWR, 0311)
	if err != nil && os.IsNotExist(err) {
		log.Warningf("Target FullchainCerts not found, create new one...")
		FullchainCerts, err = os.Create(target.FullchainCerts)
		if err != nil {
			log.Warningf("FullchainCerts create failed: %v", err)
			rollback()
			return
		}
	} else {
		log.Infof("Backing up old FullchainCerts...")
		backup, err := os.OpenFile(target.CertKey+".bak", os.O_WRONLY|os.O_CREATE, 0311)
		if err != nil {
			log.Warningf("FullchainCerts backup create failed: %v", err)
			rollback()
			return
		}
		_, err = io.Copy(backup, FullchainCerts)
		if err != nil {
			log.Warningf("FullchainCerts backup write failed: %v", err)
			rollback()
			return
		}
	}
}

func _Backup(path string, flag string) (*os.File, *os.File) {
	CertKey, err := os.OpenFile(path, os.O_RDWR, 0311)
	if err != nil && os.IsNotExist(err) {
		log.Warningf("Target " + flag + " not found, create new one...")
		CertKey, err = os.Create(path)
		if err != nil {
			log.Warningf(flag+" create failed: %v", err)
			return nil, nil
		} else {
			return CertKey, nil
		}
	} else {
		log.Infof("Backing up old " + flag + "...")
		backup, err := os.OpenFile(path+".bak", os.O_WRONLY|os.O_CREATE, 0311)
		if err != nil {
			log.Warningf("CertKey backup create failed: %v", err)
			return nil, nil
		}
		_, err = io.Copy(backup, CertKey)
		if err != nil {
			log.Warningf(flag+" backup write failed: %v", err)
			return nil, nil
		}
		return CertKey, backup
	}
}

func _Rollback(path string) error {

}
