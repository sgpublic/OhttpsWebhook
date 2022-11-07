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
	"os/exec"
	"strconv"
)

func Setup() {
	path, listen := config.GetServiceTarget()

	http.HandleFunc(path, _HandleWebhook)
	log.Infof("Start listen on http://%s%s", listen, path)
	err := http.ListenAndServe(listen, nil)
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

	// 备份 CertKey
	if _Backup(target.CertKey, "CertKey") != nil {
		return
	}
	if _Backup(target.FullchainCerts, "FullchainCerts") != nil {
		return
	}

	// 确保备份成功后，尝试写入，失败后尝试回滚
	if os.WriteFile(target.CertKey, []byte(data.Payload.CertificateCertKey), 0311) != nil {
		_ = _Rollback(target.CertKey, "CertKey")
		return
	}
	if os.WriteFile(target.FullchainCerts, []byte(data.Payload.CertificateFullchainCert), 0311) != nil {
		_ = _Rollback(target.FullchainCerts, "FullchainCerts")
		return
	}
	cmd := exec.Command("nginx", "-s", "reload")
	err = cmd.Run()
	if err != nil {
		log.Warningf("nginx reload failed: %v", err)
		return
	}
	log.Infof("Processing success! domain: %s", data.Payload.CertificateDomain)
}

func _Backup(path string, flag string) error {
	origin, err := os.OpenFile(path, os.O_RDWR, 0311)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warningf("Target %s not found, create new one...", flag)
			origin, err = os.Create(path)
			if err != nil {
				log.Warningf("%s create failed: %v", flag, err)
				return err
			} else {
				return nil
			}
		} else {
			log.Warningf("Unknown error: %v", err)
			panic(err)
			return err
		}
	} else {
		log.Infof("Backing up old %s...", flag)
		backup, err := os.OpenFile(path+".bak", os.O_WRONLY|os.O_CREATE, 0311)
		if err != nil {
			log.Warningf("CertKey backup create failed: %v", err)
			return err
		}
		_, err = io.Copy(backup, origin)
		if err != nil {
			log.Warningf("%s backup write failed: %v", flag, err)
			return err
		}
		return nil
	}
}

func _Rollback(path string, flag string) error {
	backup, err := os.OpenFile(path+".bak", os.O_RDONLY, 0311)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warningf("Backup %s not found, rollback failed!", flag)
		} else {
			log.Warningf("Unknown error: %v", err)
			panic(err)
		}
		return err
	} else {
		log.Infof("Rollbacking %s...", flag)
		origin, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0311)
		if err != nil {
			log.Warningf("CertKey backup create failed: %v", err)
			return err
		}
		_, err = io.Copy(backup, origin)
		if err != nil {
			log.Warningf("%s rollback write failed: %v", flag, err)
			return err
		}
		return nil
	}
}
