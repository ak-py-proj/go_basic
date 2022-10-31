Интерфейсы в стандартной библиотеке
В предыдущем уроке вы познакомились с интерфейсами в Go. Среди них были Stringer, Reader и Writer, уже находящиеся в стандартной библиотеке, поэтому в продакшен-код их достаточно импортировать, чтобы не писать с нуля. О других популярных интерфейсах из стандартной библиотеки расскажем ниже.
fmt.Stringer
type Stringer interface {
    String() string
} 
Этот интерфейс часто используется, когда нужно одной строчкой залогировать сложный объект. Определение интерфейса лежит в пакете fmt.
Для примера возьмём структуру User и допишем к ней реализацию интерфейса fmt.Stringer:
type User struct {
    Email        string
    PasswordHash string
    LastAccess   time.Time
}

func (u User) String() string {
    return "user with email " + u.Email
}

func main() {
    u := User{Email: "example@yandex.ru"}
    fmt.Printf("Hello, %s", u)
} 
Код выведет:
Hello, user with email example@yandex.ru 
Функция fmt.Printf использовала реализацию интерфейса.
Пакет io
Пакет io предназначен для реализации средств ввода-вывода, однако в нём есть несколько удобных интерфейсов, которые применяются и для других целей.
io.Reader
type Reader interface {
    Read(p []byte) (n int, err error)
} 
Этот интерфейс описывает чтение из любого потока данных: сети, файловой системы или буфера. Определение интерфейса лежит в пакете io.
Метод Read считывает в переданный слайс байт данные из источника. В качестве источника могут выступать любые данные, которые описаны в типе. То есть считываем их структуры и записываем в байты. Количество считанных байт неявно задаётся размером буфера — длиной слайса.
Объясним возможности интерфейса на примере. Есть буфер — и нужно прочитать байты из него. В пакете strings лежит функция strings.NewReader, которая оборачивает обычную строку в структуру strings.Reader. Эта структура имеет метод Read, значит, она реализует интерфейс io.Reader:
s := `Hodor. Hodor hodor, hodor. Hodor hodor hodor hodor hodor. Hodor. Hodor! 
Hodor hodor, hodor; hodor hodor hodor. Hodor. Hodor hodor; hodor hodor - hodor, 
hodor, hodor hodor. Hodor, hodor. Hodor. Hodor, hodor hodor hodor; hodor hodor; 
hodor hodor hodor! Hodor hodor HODOR! Hodor hodor... Hodor hodor hodor...`

// обернём строчку в strings.Reader
r := strings.NewReader(s)

// создадим буфер на 16 байт
b := make([]byte, 16)

for {
    // strings.Reader скопирует 16 байт в b
    //
    // в структуре запоминается последний указатель,  
    // то есть следующий вызов скопирует следующую порцию из 16 байт
    //
    // также метод возвращает количество прочитанных байт n и ошибку err
    //
    // когда дойдём до конца строки, метод отдаст ошибку io.EOF
    n, err := r.Read(b)

    // при работе с интерфейсом io.Reader нужно в первую очередь проверять
    // n > 0, затем err != nil
    //
    // могут быть ситуации, когда часть данных получилось прочитать
    // и сохранить в буфер, а затем произошла ошибка 
    //
    // в таком случае будут одновременно n > 0 и err != nil
    if n > 0 {
        // выведем на экран содержимое буфера
        fmt.Printf("%v\n", b)
    }

    if err != nil {
        // если дочитали до конца, выходим из цикла
        if errors.Is(err, io.EOF) {
            break
        }

        // обрабатываем ошибку чтения
        fmt.Printf("error: %v\n", err)
    }
} 
Удобство применения io.Reader в том, что его пользователь может вообще не знать, откуда берутся данные: из файла, сети или генерируются на лету. Интерфейс описывает унифицированный метод работы с ними.
Для закрепления реализуем генератор случайных данных:
package randbyte

import (
    "io"
    "math/rand"
)

type generator struct {
    rnd rand.Source // Генератор случайных чисел. Вообще rand.Rand уже реализует интерфейс io.Reader, но для примера мы реализуем его самостоятельно.
}

// New — обратите внимание, что мы возвращаем generator, присвоенный интерфейсу io.Reader, сама структура generator неэкспортируемая.
// Мы скрыли внутри пакета все детали.
func New(seed int64) io.Reader {
    return &generator{
        rnd: rand.NewSource(seed),
    }
}

// Read — реализация io.Reader
func (g *generator) Read(bytes []byte) (n int, err error) { // error — это тип ошибки, подробнее мы рассмотрим его в следующем разделе.
    for i := range bytes {
        randInt := g.rnd.Int63()  // функция возвращает положительное число в пределах от 0 до 2^63
        randByte := byte(randInt) // приводим к типу byte
        bytes[i] = randByte
    }
    return len(bytes), nil
}
 
package main

import (
    "example/randbyte"
    "fmt"
    "time"
)

func main() {

    // создаём генератор случайных чисел
    generator := randbyte.New(time.Now().UnixNano()) // в качестве затравки передаём ему текущее время, и при каждом запуске оно будет разным.

    buf := make([]byte, 16)

    for i := 0; i < 5; i++ {
        n, _ := generator.Read(buf) // единственный доступный метод, но он нам и нужен.
        fmt.Printf("Generate bytes: %v size(%d)\n", buf, n)
    }

} 
Мы реализовали простой генератор случайных байт.
Задание 1 из 2
В последнем рассмотренном примере реализация функции Read не очень эффективна — генератор случайных чисел возвращает 64-битное число, то есть 8 байт. Из них используем только 1.
Попробуйте реализовать более эффективное решение. Для упрощения примера считайте, что функция будет принимать только слайсы, длина которых кратна 8. Для преобразования числа в слайс байт можно использовать функцию из стандартной библиотеки binary.LittleEndian.PutUint64([ ]byte, uint64).
Готовы проверить себя?


Правильный ответ
Да
// Read — реализация io.Reader
func (g *generator) Read(bytes []byte) (n int, err error) { // error это тип ошибки, подробнее мы рассмотрим его в следующем разделе.
    for i := 0; i+8 < len(bytes); i += 8 {
        binary.LittleEndian.PutUint64(bytes[i:i+8], uint64(g.rnd.Int63()))
    }
    return len(bytes), nil
} 
io.Writer
type Writer interface {
    Write(p []byte) (n int, err error)
} 
Этот интерфейс означает запись в любой возможный поток данных: сетевой сокет, файл или буфер. Определение интерфейса лежит в пакете io.
C этим интерфейсом ситуация, обратная io.Reader. Он позволяет записать переданный ему слайс байт куда-то. Куда именно — определяется реализацией.
Для примера соберём большую строку из подстрок, вот только не через оператор +=, потому что тогда на каждую итерацию будет лишняя копия всей строки. В пакете strings есть структура strings.Builder для сборки строки без избыточного копирования. Эта структура имеет метод Write, значит, она реализует интерфейс io.Writer:
// создаём strings.Builder
w := strings.Builder{}

for i := 0; i < 50; i++ {
    // функция fmt.Fprintf принимает аргументом io.Writer
    // благодаря этому можно записывать форматированный вывод
    fmt.Fprintf(&w, "%v", math.NaN())
}

w.Write([]byte("... BATMAN!"))

// выводим собранную строку
fmt.Printf("%s\n", &w) 
Приведём пример реализации интерфейса Write. Предположим, что мы хотим посчитать хеш от некоторого массива байт или наборов массивов. Для простоты возьмём упрощённую функцию хеширования:
package hashbyte

import "io"

type Hasher interface {
    io.Writer // мы встроили интерфейс io.Writer в наш интерфейс, чтобы задать требование по наличию метода Write
    Hash() byte
}

type hash struct {
    result byte
}

func New(_init byte) Hasher {
    return &hash{
        result: _init,
    }
}

// Write — сюда может быть записан массив байт любой длины, для которой будет подсчитываться хэш.
func (h *hash) Write(bytes []byte) (n int, err error) {
    // обновляем хеш для каждого байта, записанного в хешер
    for _, b := range bytes {
        h.result = (h.result^b)<<1 + b%2 
    }
    return len(bytes), nil
}

func (h hash) Hash() byte {
    return h.result
} 
Теперь используем её в нашей программе:
func main() {

    // создаём генератор случайных чисел
    generator := randbyte.New(time.Now().UnixNano()) // в качестве затравки передаём ему текущее время — при каждом запуске оно будет разным

    buf := make([]byte, 16)

    for i := 0; i < 5; i++ {
        n, _ := generator.Read(buf)
        fmt.Printf("Generate bytes: %v size(%d)\n", buf, n)
    }

    hasher := hashbyte.New(0)
    hasher.Write(buf)
    fmt.Printf("Hash: %v \n", hasher.Hash())

}
 
Функции-утилиты для io.Reader и io.Writer
io.Copy
func Copy(dst Writer, src Reader) (written int64, err error) 
Функция копирует все байты из io.Reader в io.Writer.
Данные будут считываться до тех пор, пока функция Read не вернёт вторым аргументом ошибку. Если в качестве ошибки будет возвращено значение io.EOF, то выполнение функции закончится без ошибок. Также будет возвращено количество байт.
io.EOF происходит от end of frame (конец файла) — исторически так назывался специальный символ, который означал конец файла.
Приведём простой пример. Напишем функцию, копирующую содержимое одного файла в другой:

func CopyFile(srcFileName, dstFileName string) error {
    srcFile, err := os.Open(srcFileName)
    if err != nil {
        return err
    }
    dstFile, err := os.Create(dstFileName)
    if err != nil {
        return err
    }
    n, err := io.Copy(dstFile, srcFile)
    if err != nil {
        return err
    }
    fmt.Printf("Copied %d bytes from %s to %s", n, srcFileName, dstFileName)
    return nil
}
 
Структура типа os.File реализует интерфейсы io.Reader и io.Writer.
Было бы просто считать весь исходный файл в память и затем скопировать его в новый. Но если исходный файл занимает сотни гигабайт? io.Copy работает умнее, считывая и записывая данные небольшими кусочками, поэтому для подобных операций рекомендуется использовать именно её.
io.CopyN
func CopyN(dst Writer, src Reader, n int64) (written int64, err error) 
Функция копирует все байты из io.Reader в io.Writer, но не более n байт. То же самое, что и Copy, но с ограничением — можно использовать с источниками данных, которые слишком большие или вообще бесконечные. Например, напишем функцию, которая будет сохранять данные из нашего генератора случайных чисел в файл.
// Dump — сохраняет вычисленные данные в файл
func (g generator) Dump(n int64, dst *os.File) error {
    _, err := io.CopyN(g, dst, n)
    return err
} 
Если бы мы использовали Copy, то программа продолжила бы работать до переполнения диска.
io.ReadAll
func ReadAll(r Reader) ([]byte, error) 
Функция считывает все байты из io.Reader. Чтение закончится, когда io.Reader вернёт io.EOF.
io.ReadAtLeast
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) 
Функция считывает байты из io.Reader c ограничением: если прочитанных байт оказалось меньше, чем n, вернётся ошибка io.ErrUnexpectedEOF. Это используется при парсинге бинарных данных, чтобы гарантировать, что нужное минимальное количество байт будет вычитано.
Другие интерфейсы пакета io
Мы привели примеры основных методов работы с функциями и интерфейсами ввода-вывода. Кроме них, в пакете осталось ещё много интересного. Рекомендуем открыть документацию пакета io, чтобы посмотреть на определения остальных интерфейсов. io.Reader и io.Writer — основные интерфейсы, но могут пригодиться и другие.
Задание 2
В пакете io есть функция LimitReader(r io.Reader, n int64) io.Reader. Она ограничивает количество байт, которое можно вычитать из io.Reader.
Запрограммируйте подобную функцию самостоятельно:
package main

import (
    "io"
    "log"
    "os"
    "strings"
)

func LimitReader(r io.Reader, n int) io.Reader {
    // ...
}

func main() {
    r := strings.NewReader("some io.Reader stream to be read\n")
    lr := LimitReader(r, 4)

    _, err := io.Copy(os.Stdout, lr)
    if err != nil {
        log.Fatal(err)
    }
} 
Код должен вывести подстроку some.
Подсказка: подумайте, как можно ограничить чтение. Нужно как-то подсчитывать и запоминать количество байт, оставшихся для чтения из reader.

package main

import (
    "io"
    "log"
    "os"
    "strings"
)

type LimitedReader struct {
    reader io.Reader
    //  запоминаем количество считанных байт
    left   int
}

func LimitReader(r io.Reader, n int) io.Reader {
    return &LimitedReader{reader: r, left: n}
}

func (r *LimitedReader) Read(p []byte) (int, error) {
    if r.left == 0 {
        return 0, io.EOF
    }
    if r.left < len(p) {
        p = p[0:r.left]
    }
    n, err := r.reader.Read(p)
    r.left -= n
    return n, err
}

func main() {
    r := strings.NewReader("some io.Reader stream to be read\n")
    lr := LimitReader(r, 4)

    _, err := io.Copy(os.Stdout, lr)
    if err != nil {
        log.Fatal(err)
    }
}
