package crypto

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ouyangzhongmin/nano/pipeline"
	"github.com/ouyangzhongmin/nano/session"
	"github.com/xxtea/xxtea-go/xxtea"
)

var xxteaKey = []byte("7AEC4MA152BQE9HWQ7KB")

type Crypto struct {
	key []byte
}

func NewCrypto() *Crypto {
	return &Crypto{xxteaKey}
}

func (c *Crypto) Inbound(s *session.Session, msg *pipeline.Message) error {
	out, err := base64.StdEncoding.DecodeString(string(msg.Data))
	if err != nil {
		fmt.Printf("Inbound Error=%s, In=%s \n", err.Error(), string(msg.Data))
		return err
	}

	out = xxtea.Decrypt(out, c.key)
	if out == nil {
		return errors.New("decrypt error: out == nil")
	}
	msg.Data = out
	return nil
}

func (c *Crypto) Outbound(s *session.Session, msg *pipeline.Message) error {
	out := xxtea.Encrypt(msg.Data, c.key)
	msg.Data = []byte(base64.StdEncoding.EncodeToString(out))
	return nil
}
