package password_test

import (
	"testing"

	"github.com/alex023/basekit/password"
)

func Test_ReturnPassword(t *testing.T) {
	passFirst := "测试的明文密码,但不知道需要输入多少位"

	salt, hash1 := password.ReturnPassword(passFirst)
	hash2 := password.GenerateHash(salt, passFirst)

	if hash1 != hash2 {
		t.Error("ERROR！两次结果不匹配。")
	}
}

//测试DesDecrypt与DesEncrypt是否能够匹配，输入值采用随机方式生成。
func Test_DesDecryptEncrypt(t *testing.T) {
	//①初始化一个判断值
	result := true
	//②循环多次执行
	for i := 0; i < 5000; i++ {
		key := []byte(password.GenerateSalt(8, 31))
		//		origData := []byte("沧海一声笑,涛涛两岸潮")
		origData := []byte(password.GenerateSalt(16, 29))
		result = desDecryptEncrypt(t, key, origData)
		//③result为false，说明加解密算法不匹配，则跳出
		if result == false {
			break
		}
	}
	if !result {
		t.Error("ERROR！DesDecrypt与DesEncrypt无法匹配。")
	}

}

//针对具体的key，检测加解密的匹配程度。如果匹配，则返回true，否则为false.
func desDecryptEncrypt(t *testing.T, key []byte, origData []byte) bool {
	//①加密
	cryptedData, err := password.DesEncrypt(origData, key)
	if err != nil {

	}
	//②解密
	tempOrigData, err := password.DesDecrypt(cryptedData, key)
	//③ 判断输入值，与加解密之后的值是否相同
	if string(origData) == string(tempOrigData) {
		return true
	} else {
		return false
	}
}

// 可能用到的序号：①②③④⑤⑥⑦⑧⑨
