package producer

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func GenerateRandomColor() color.RGBA {
	// 真随机数,如果我们的应用对安全性要求比较高，需要使用真随机数的话，那么可以使用 crypto/rand 包中的方法,这样生成的每次都是不同的随机数.
	// 生成随机颜色
	result1, _ := rand.Int(rand.Reader, big.NewInt(256))
	result2, _ := rand.Int(rand.Reader, big.NewInt(256))
	result3, _ := rand.Int(rand.Reader, big.NewInt(256))
	return color.RGBA{
		uint8(result1.Uint64()),
		uint8(result2.Uint64()),
		uint8(result3.Uint64()),
		255, // 不透明
	}
}

func GenerateInitialImage(nameStr string) (string, error) {

	// 从用户名中获取首字母
	initial := string([]rune(nameStr)[0])

	// 创建一个新的RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// 设置背景颜色为随机颜色
	bgColor := GenerateRandomColor()
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 在图像上绘制文字，调整文字大小
	face := basicfont.Face7x13

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{0, 0, 0, 255}),
		Face: face,
		Dot:  fixed.P(50, 50), // 调整文本位置
	}
	d.DrawString(initial)

	// 编码图像为PNG格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	// 将图像数据转为Base64字符串
	base64Image := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))

	return base64Image, nil
}
