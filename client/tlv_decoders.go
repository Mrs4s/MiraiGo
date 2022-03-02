package client

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Mrs4s/MiraiGo/utils"

	"github.com/Mrs4s/MiraiGo/binary"
)

// --- tlv decoders for qq client ---

/*
func (c *QQClient) decodeT161(data []byte) {
	reader := binary.NewReader(data)
	reader.ReadBytes(2)
	t := reader.ReadTlvMap(2)
	if t172, ok := t[0x172]; ok {
		c.rollbackSig = t172
	}
}
*/

func (c *QQClient) decodeT119(data, ek []byte) {
	tea := binary.NewTeaCipher(ek)
	reader := binary.NewReader(tea.Decrypt(data))
	reader.ReadBytes(2)
	m := reader.ReadTlvMap(2)
	if t130, ok := m[0x130]; ok {
		c.decodeT130(t130)
	}
	if t113, ok := m[0x113]; ok {
		c.decodeT113(t113)
	}
	/*
		if t528, ok := m[0x528]; ok {
			c.t528 = t528
		}
		if t530, ok := m[0x530]; ok {
			c.t530 = t530
		}
	*/
	if t108, ok := m[0x108]; ok {
		c.sig.Ksid = t108
	}

	var (
		// openId   []byte
		// openKey  []byte
		// payToken []byte
		// pf       []byte
		// pfkey    []byte
		gender uint16 = 0
		age    uint16 = 0
		nick          = ""
		// a1       []byte
		// noPicSig []byte
		// ctime           = time.Now().Unix()
		// etime           = ctime + 2160000
		psKeyMap    map[string][]byte
		pt4TokenMap map[string][]byte
	)

	if t11a, ok := m[0x11a]; ok {
		nick, age, gender = readT11A(t11a)
	}
	/*
		if _, ok := m[0x125]; ok {
			openId, openKey = readT125(t125)
		}
		if t186, ok := m[0x186]; ok {
		 	c.decodeT186(t186)
		}
		if _, ok := m[0x199]; ok {
			openId, payToken = readT199(t199)
		}
		if _, ok := m[0x200]; ok {
			pf, pfkey = readT200(t200)
		}
		if _, ok := m[0x531]; ok {
			a1, noPicSig = readT531(t531)
		}
		if _, ok := m[0x138]; ok {
			readT138(t138) // chg time
		}
	*/
	if t512, ok := m[0x512]; ok {
		psKeyMap, pt4TokenMap = readT512(t512)
	}
	c.oicq.WtSessionTicketKey = utils.Select(m[0x134], c.oicq.WtSessionTicketKey)

	// we don't use `c.sigInfo = &auth.SigInfo{...}` here,
	// because we need keep other fields in `c.sigInfo`
	s := c.sig
	s.LoginBitmap = 0
	s.SrmToken = utils.Select(m[0x16a], s.SrmToken)
	s.T133 = utils.Select(m[0x133], s.T133)
	s.EncryptedA1 = utils.Select(m[0x106], s.EncryptedA1)
	s.TGT = m[0x10a]
	s.TGTKey = m[0x10d]
	s.UserStKey = m[0x10e]
	s.UserStWebSig = m[0x103]
	s.SKey = m[0x120]
	s.SKeyExpiredTime = time.Now().Unix() + 21600
	s.D2 = m[0x143]
	s.D2Key = m[0x305]
	s.DeviceToken = m[0x322]

	s.PsKeyMap = psKeyMap
	s.Pt4TokenMap = pt4TokenMap
	if len(c.PasswordMd5[:]) > 0 {
		data, cl := binary.OpenWriterF(func(w *binary.Writer) {
			w.Write(c.PasswordMd5[:])
			w.WriteUInt32(0) // []byte{0x00, 0x00, 0x00, 0x00}...
			w.WriteUInt32(uint32(c.Uin))
		})
		key := md5.Sum(data)
		cl()
		decrypted := binary.NewTeaCipher(key[:]).Decrypt(c.sig.EncryptedA1)
		if len(decrypted) > 51+16 {
			dr := binary.NewReader(decrypted)
			dr.ReadBytes(51)
			c.deviceInfo.TgtgtKey = dr.ReadBytes(16)
		}
	}
	c.Nickname = nick
	c.Age = age
	c.Gender = gender
}

// wtlogin.exchange_emp
func (c *QQClient) decodeT119R(data []byte) {
	tea := binary.NewTeaCipher(c.deviceInfo.TgtgtKey)
	reader := binary.NewReader(tea.Decrypt(data))
	reader.ReadBytes(2)
	m := reader.ReadTlvMap(2)
	if t120, ok := m[0x120]; ok {
		c.sig.SKey = t120
		c.sig.SKeyExpiredTime = time.Now().Unix() + 21600
		c.debug("skey updated: %v", c.sig.SKey)
	}
	if t11a, ok := m[0x11a]; ok {
		c.Nickname, c.Age, c.Gender = readT11A(t11a)
		c.debug("account info updated: " + c.Nickname)
	}
}

func (c *QQClient) decodeT130(data []byte) {
	reader := binary.NewReader(data)
	reader.ReadBytes(2)
	// c.timeDiff = int64(reader.ReadInt32()) - time.Now().Unix()
	// c.t149 = reader.ReadBytes(4)
}

func (c *QQClient) decodeT113(data []byte) {
	reader := binary.NewReader(data)
	uin := reader.ReadInt32() // ?
	fmt.Println("got t113 uin:", uin)
}

/*
func (c *QQClient) decodeT186(data []byte) {
	// c.pwdFlag = data[1] == 1
}
*/

// --- tlv readers ---

func readT125(data []byte) (openID, openKey []byte) {
	reader := binary.NewReader(data)
	openID = reader.ReadBytesShort()
	openKey = reader.ReadBytesShort()
	return
}

func readT11A(data []byte) (nick string, age, gender uint16) {
	reader := binary.NewReader(data)
	reader.ReadUInt16()
	age = uint16(reader.ReadByte())
	gender = uint16(reader.ReadByte())
	nick = reader.ReadStringLimit(int(reader.ReadByte()) & 0xff)
	return
}

func readT199(data []byte) (openID, payToken []byte) {
	reader := binary.NewReader(data)
	openID = reader.ReadBytesShort()
	payToken = reader.ReadBytesShort()
	return
}

func readT200(data []byte) (pf, pfKey []byte) {
	reader := binary.NewReader(data)
	pf = reader.ReadBytesShort()
	pfKey = reader.ReadBytesShort()
	return
}

func readT512(data []byte) (psKeyMap map[string][]byte, pt4TokenMap map[string][]byte) {
	reader := binary.NewReader(data)
	length := int(reader.ReadUInt16())

	psKeyMap = make(map[string][]byte, length)
	pt4TokenMap = make(map[string][]byte, length)

	for i := 0; i < length; i++ {
		domain := reader.ReadStringShort()
		psKey := reader.ReadBytesShort()
		pt4Token := reader.ReadBytesShort()

		if len(psKey) > 0 {
			psKeyMap[domain] = psKey
		}

		if len(pt4Token) > 0 {
			pt4TokenMap[domain] = pt4Token
		}
	}

	return
}

func readT531(data []byte) (a1, noPicSig []byte) {
	reader := binary.NewReader(data)
	m := reader.ReadTlvMap(2)
	if m.Exists(0x103) && m.Exists(0x16a) && m.Exists(0x113) && m.Exists(0x10c) {
		a1 = append(m[0x106], m[0x10c]...)
		noPicSig = m[0x16a]
	}
	return
}
