package main

import (
    "os"
    "github.com/codegangsta/cli"
    "bufio"	
	"fmt"
    "strconv"
    "time"
    "strings"
    "math/rand"    
    "github.com/nimiri/go-shnmk16"
    "github.com/nimiri/go-jisx0208"
    tm "github.com/buger/goterm"
)

func main() {
    app := cli.NewApp()
    app.Name = "tcho"            // ヘルプを表示する際に使用される
    app.Usage = "print tsuyoi esho" // ヘルプを表示する際に使用される
    app.Version = "0.0.1"         // ヘルプを表示する際に使用される
    app.Action = func(c *cli.Context) { // コマンド実行時の処理

        println("Number of arguments is", len(c.Args()))

        if len(c.Args()) == 0 {
            println("[ERROR] No Arguments!")
            return
        }
        input := strings.Join(c.Args(), " ")

        inputRune := []rune(input)
        lenInputStr := len(inputRune)
    
        // コマンドの引数をjisx0208のコードに変換
        // 文字列を渡して、コードの配列を取得
        codes := []int{}
        for i := 0; i < lenInputStr; i++ {
            
            var code int
            var err error
            if int(inputRune[i]) <= 255 {
                code = int(inputRune[i])
            } else { 
                code, err = jisx0208.Code(inputRune[i])
                if err != nil {
                    panic(err)
                }
            }

            codes = append(codes, code)
        }

        // フォントファイルの先頭からのoffsetを取得
        // コードの配列を渡して、ファイルポインタの先頭からoffsetするbyte数の配列を取得
        offsets := []int{}
        for i := 0; i < lenInputStr; i++ {
            offset, err := shnmk16.Offset(codes[i])
            if err != nil {
                    panic(err)
            }

            offsets = append(offsets, offset)
        }
            
        // フォントのファイルを開く
        fpFull, err := os.Open("./assets/shnmk16.bdf")
        if err != nil {
            panic(err)
        }
        defer fpFull.Close()
        fpHalf, err := os.Open("./assets/shnm8x16.bdf")
        if err != nil {
            panic(err)
        }
        defer fpHalf.Close()

        // offsetの配列を渡して、Bitmapの配列を取得
        bitmaps := []string{}
        var fp *os.File
        for i := 0; i < lenInputStr; i++ {

            if codes[i] < 255 {
                fp = fpHalf
            } else {
                fp = fpFull
            }

            fp.Seek(int64(offsets[i]), 0)
            scanner := bufio.NewScanner(fp)

            // 配置部分までファイルポインタを進める
            for {
                scanner.Scan()
                scanRow := scanner.Text()
                if err := scanner.Err(); err != nil {
                    panic(err)
                }

                if scanRow == "BITMAP" {
                    break
                }
            }

            j := 0
            for {
                scanner.Scan()
                scanRow := scanner.Text()
                if err := scanner.Err(); err != nil {
                    panic(err)
                }

                if scanRow == "ENDCHAR" {
                    break
                }

                // 1文字目の場合、前に3文字分のスペースを追加
                if i == 0 {
                    bitmaps = append(bitmaps, strings.Repeat("0", 48))
                }

                // 16進数の文字列を2進数にして0詰めで連結
                num, _ := strconv.ParseInt(scanRow, 16, 32)
                bitmaps[j] = bitmaps[j] + "00" + fmt.Sprintf("%016b",num)

                j++
            }
            
        }

        // Bitmapの配列を取得を渡して、いい感じのバイナリっぽい文字列を取得
        // 表示に使用する文字
        displayStringArray := []string{"1", "2", "3", "4", "5", "6", "7", "A", "E", "F"}
        allStringArray := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}
        zerosString := strings.Repeat("0", 50) 
        zerosStringArray := strings.Split(zerosString, "")

        // 程よく間を開けるため、allStringArrayに 0 をたくさん増やして、 0の比率を増やす
        paddingStringArray := append(allStringArray, zerosStringArray...)

        lenBitmapRow := len(bitmaps) - 1
        lenBitmapCol := len(bitmaps[0])
    
        // 元の配列をいじらないように、表示領域用の
        // displayBitmap := []string{}
        for i := 0; i < lenBitmapCol; i++ {
            tm.Clear()
            tm.MoveCursor(1,1)

            rand.Seed(time.Now().UnixNano()) 
            // 画面高さ（実行中に画面サイズが変わった時に備えて、ループ内で取得）
            consoleHeight := tm.Height()
            // 画面幅（実行中に画面サイズが変わった時に備えて、ループ内で取得）
            consoleWidth := tm.Width()

            // 上下のpaddingに必要な行数
            paddingRow := (consoleHeight - lenBitmapRow) / 2

            // 画面中央に文字を表示したいので、上のpadding表示 
            for j := 0; j < paddingRow; j++ {
                for k := 0; k < consoleWidth; k++ {
                    // ランダムに文字列を生成して埋める
                    tm.Printf(paddingStringArray[rand.Intn(len(paddingStringArray))])
                }
                tm.Println()
            }

            // ホントの表示部分
            for j := 0; j < lenBitmapRow; j++ {
                for k := i; k < consoleWidth + i; k++ {
                    // 表示
                    if k < lenBitmapCol && bitmaps[j][k] == '1'{
                        // 値を出す部分は、ランダムに文字列を生成して埋める
                        tm.Printf(displayStringArray[rand.Intn(len(displayStringArray))])
                    } else {
                        // 空白部分は 0 多めの文字列をランダムで出力
                        tm.Printf("%v", paddingStringArray[rand.Intn(len(paddingStringArray))])
                    }
                }
                tm.Println()
            }

            // 下のpadding表示
            for j := 0; j < paddingRow; j++ {
                for k := 0; k < consoleWidth; k++ {
                    // ランダムに文字列を生成して埋める
                    tm.Printf(paddingStringArray[rand.Intn(len(paddingStringArray))])
                }
                tm.Println()
            }

            // 画面リフレッシュ
            tm.Flush()
            // 早すぎると一瞬で流れきるので、待機
            time.Sleep(time.Second/25)
        }

        tm.Clear()
        tm.MoveCursor(0,0)        
        tm.Flush()
    }
    app.Run(os.Args)
}

var randSrc = rand.NewSource(time.Now().UnixNano())

const (
    rs6Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    rs6LetterIdxBits = 6
    rs6LetterIdxMask = 1<<rs6LetterIdxBits - 1
    rs6LetterIdxMax = 63 / rs6LetterIdxBits
)

func RandString(n int) string {
    b := make([]byte, n)
    cache, remain := randSrc.Int63(), rs6LetterIdxMax
    for i := n-1; i >= 0; {
        if remain == 0 {
            cache, remain = randSrc.Int63(), rs6LetterIdxMax
        }
        idx := int(cache & rs6LetterIdxMask)
        if idx < len(rs6Letters) {
            b[i] = rs6Letters[idx]
            i--
        }
        cache >>= rs6LetterIdxBits
        remain--
    }
    return string(b)
}