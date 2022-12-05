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

func Setup() {
	path, listen := config.GetServiceTarget()

	http.HandleFunc(path, _HandleWebhook)
	log.Infof("Starting listen on http://%s%s", listen, path)
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
	if r.Method != "POST" {
		log.Warningf("Unsupported method: %s", r.Method)
		return
	}

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
		CertificateName          string   `json:"certificateName"`
		CertificateDomains       []string `json:"certificateDomains"`
		CertificateCertKey       string   `json:"certificateCertKey"`
		CertificateFullchainCert string   `json:"certificateFullchainCerts"`
	} `json:"payload"`
	Sign string `json:"sign"`
}

func _Process(data Ohttps) {
	log.Infof("Processing request, CertID: %s, sign: %s", data.Payload.CertificateName, data.Sign)
	for _, domain := range data.Payload.CertificateDomains {
		log.Infof("Processing domain: %s", domain)
		target, err := config.GetTarget(domain)
		if err != nil {
			log.Warningf("Error occupied during the process (domain: %s): %v", domain, err)
			return
		}

		// 校验 sign
		h := md5.New()
		h.Write([]byte(strconv.Itoa(data.Timestamp) + ":" + config.GetWebhookKey()))
		sign := hex.EncodeToString(h.Sum(nil))
		if data.Sign != sign {
			log.Warningf("Sign mismatch! domain: %s", domain)
			return
		}

		// 备份 CertKey
		if _Backup(target.CertKey, "CertKey", domain) != nil {
			return
		}
		if _Backup(target.FullchainCerts, "FullchainCerts", domain) != nil {
			return
		}

		// 确保备份成功后，尝试写入，失败后尝试回滚
		log.Infof("Writting new CertKey... (domain: %s)", domain)
		err = os.WriteFile(target.CertKey, []byte(data.Payload.CertificateCertKey), 0644)
		if err != nil {
			log.Warningf("CertKey writing failed (domain: %s): %v", domain, err)
			_ = _Rollback(target.CertKey, "CertKey")
			return
		}
		log.Infof("Writting new FullchainCerts... (domain: %s)", domain)
		err = os.WriteFile(target.FullchainCerts, []byte(data.Payload.CertificateFullchainCert), 0644)
		if err != nil {
			log.Warningf("FullchainCerts writing failed (domain: %s): %v", domain, err)
			_ = _Rollback(target.FullchainCerts, "FullchainCerts")
			return
		}
	}
	err := config.GetNginxReloadCommand().Run()
	if err != nil {
		log.Warningf("nginx reload failed: %v", err)
		return
	}
	log.Infof("Processing success! CertID: %s", data.Payload.CertificateName)
}

func _Backup(path string, flag string, domain string) error {
	log.Infof("Backing up %s... (domain: %s)", flag, domain)
	origin, err := os.OpenFile(path, os.O_RDWR, 0644)
	//goland:noinspection GoUnhandledErrorResult
	defer origin.Close()
	if err != nil {
		if os.IsNotExist(err) {
			log.Warningf("Target %s not found, create new one... (domain: %s)", flag, domain)
			origin, err = os.Create(path)
			if err != nil {
				log.Warningf("%s create failed (domain: %s): %v", flag, domain, err)
				return err
			} else {
				return nil
			}
		} else {
			log.Warningf("Unknown error (domain: %s): %v", domain, err)
			panic(err)
			return err
		}
	} else {
		backup, err := os.OpenFile(path+".bak", os.O_WRONLY|os.O_CREATE, 0644)
		//goland:noinspection GoUnhandledErrorResult
		defer backup.Close()
		if err != nil {
			log.Warningf("CertKey backup create failed (domain: %s): %v", domain, err)
			return err
		}
		_, err = io.Copy(backup, origin)
		if err != nil {
			log.Warningf("%s backup write failed (domain: %s): %v", flag, domain, err)
			return err
		}
		return nil
	}
}

func _Rollback(path string, flag string) error {
	log.Infof("Rollbacking %s...", flag)
	backup, err := os.OpenFile(path+".bak", os.O_RDONLY, 0644)
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
		origin, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
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
