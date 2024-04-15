package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

type Elemento struct {
	simbolo  rune
	cor      termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}

var personagem = Elemento{
	simbolo:  '☺',
	cor:      termbox.ColorCyan,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var parede = Elemento{
	simbolo:  '▣',
	cor:      termbox.ColorRed,
	corFundo: termbox.ColorBlue,
	tangivel: true,
}

var vegetacao = Elemento{
	simbolo:  '♣',
	cor:      termbox.ColorGreen,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var inimigo = Elemento{
	simbolo:  '▶',
	cor:      termbox.ColorLightRed,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var moeda = Elemento{
	simbolo:  '◉',
	cor:      termbox.ColorYellow,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var portal = Elemento{
	simbolo:  '⚑',
	cor:      termbox.ColorMagenta,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var vitoria = Elemento{
	simbolo:  '★',
	cor:      termbox.ColorCyan,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var vazio = Elemento{
	simbolo:  ' ',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var neblina = Elemento{
	simbolo:  '.',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorYellow,
	tangivel: false,
}

var estrelaPosicionada bool
var piscarMensagem bool

var moedasColetadasMap = make(map[int]bool)

var mapa [][]Elemento
var posX, posY int
var posXInicial, posYInicial int
var ultimoElementoSobPersonagem = vazio
var statusMsg string

var efeitoNeblina = false
var revelado [][]bool
var raioVisao int = 3

var totalDeMoedas = 10
var moedasColetadas int
var vidas = 5
var mutex sync.Mutex

var startTime time.Time
var jogoEmAndamento bool

var exibindoAvisoMoedas bool
var tempoAvisoMoedas time.Time

var vegetacaoVariacoes = []rune{'♣', '♧'}
var continuarMovimentacaoVegetacao bool

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	StartGame()
	carregarMapa("mapa.txt")
	if efeitoNeblina {
		revelarArea()
	}
	mapa[posY][posX] = personagem
	desenhaTudo()
	go moverInimigo()
	go moverVegetacao()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return
			}
			if ev.Ch == 'e' {
				interagir()
			} else {
				mover(ev.Ch)
			}
			desenhaTudo()
		}
		termbox.Flush()
	}
}

func carregarMapa(nomeArquivo string) {
	arquivo, err := os.Open(nomeArquivo)
	if err != nil {
		panic(err)
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	y := 0
	for scanner.Scan() {
		linhaTexto := scanner.Text()
		var linhaElementos []Elemento
		var linhaRevelada []bool
		for x, char := range linhaTexto {
			elementoAtual := vazio
			switch char {
			case parede.simbolo:
				elementoAtual = parede
			case moeda.simbolo:
				elementoAtual = moeda
			case vegetacao.simbolo:
				elementoAtual = vegetacao
			case inimigo.simbolo:
				elementoAtual = inimigo
			case portal.simbolo:
				elementoAtual = portal
			case vitoria.simbolo:
				elementoAtual = vitoria
			case neblina.simbolo:
				elementoAtual = neblina
			case personagem.simbolo:
				posX, posY = x, y
				posXInicial, posYInicial = x, y
				elementoAtual = vazio
			}
			linhaElementos = append(linhaElementos, elementoAtual)
			linhaRevelada = append(linhaRevelada, false)
		}
		mapa = append(mapa, linhaElementos)
		revelado = append(revelado, linhaRevelada)
		y++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y, linha := range mapa {
		for x, elem := range linha {
			if efeitoNeblina == false || revelado[y][x] {
				if elem == moeda && moedasColetadasMap[y*100+x] {
					elem = vazio
				}
				termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
			} else {
				termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
			}
		}
	}

	desenhaBarraDeStatus()

	termbox.Flush()
}

func desenhaBarraDeStatus() {
	tempoDeJogo := time.Since(startTime).Round(time.Second)
	tempoMsg := fmt.Sprintf("Tempo de jogo: %s", tempoDeJogo)

	if jogoEmAndamento {
		for i, c := range tempoMsg {
			termbox.SetCell(i, len(mapa)+3, c, termbox.ColorLightBlue, termbox.ColorDefault)
		}
	}

	msg := "Use WASD para mover e E para interagir. ESC para sair. Moedas coletadas: " + fmt.Sprintf("%d", moedasColetadas) + " Vidas: " + fmt.Sprintf("%d", vidas)
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+2, c, termbox.ColorLightBlue, termbox.ColorDefault)
	}
}

func revelarArea() {
	minX := max(0, posX-raioVisao)
	maxX := min(len(mapa[0])-1, posX+raioVisao)
	minY := max(0, posY-raioVisao/2)
	maxY := min(len(mapa)-1, posY+raioVisao/2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {

			revelado[y][x] = true
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func moverInimigo() {
	var direcao int = 1
	var movimentos int = 0
	const maxMovimentos = 2

	for {
		time.Sleep(500 * time.Millisecond)

		mutex.Lock()
		var inimigoX, inimigoY int
		encontrado := false
		for y, linha := range mapa {
			for x, char := range linha {
				if char == inimigo {
					inimigoX, inimigoY = x, y
					encontrado = true
					break
				}
			}
			if encontrado {
				break
			}
		}

		if !encontrado {
			mutex.Unlock()
			continue
		}

		novaY := inimigoY + direcao
		if movimentos >= maxMovimentos || novaY < 0 || novaY >= len(mapa) || mapa[novaY][inimigoX] == parede {
			direcao *= -1
			movimentos = 0
		} else {
			movimentos++
			if novaY == posY && inimigoX == posX {
				vidas--

				termbox.SetCell(posX, posY, personagem.simbolo, termbox.ColorDefault, termbox.ColorRed)
				termbox.Flush()
				time.Sleep(200 * time.Millisecond)

				if posX > 0 && mapa[posY][posX-1] == vazio {
					mapa[posY][posX] = vazio
					posX--
					ultimoElementoSobPersonagem = vazio
					mapa[posY][posX] = personagem
				}

				if vidas == 0 {
					GameOver()
					mutex.Unlock()
					return
				}
			}

			mapa[inimigoY][inimigoX] = vazio
			mapa[novaY][inimigoX] = inimigo
			inimigoY = novaY
		}
		mutex.Unlock()
		desenhaTudo()
	}
}

func mover(comando rune) {
	dx, dy := 0, 0
	switch comando {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}
	novaPosX, novaPosY := posX+dx, posY+dy
	if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) {
		if mapa[novaPosY][novaPosX].tangivel && mapa[novaPosY][novaPosX] != portal {
			if mapa[novaPosY][novaPosX] == inimigo {
				vidas--
				EfeitoDano(novaPosX, novaPosY)

				if vidas == 0 {
					GameOver()
				}
				return
			}
			return
		}

		if mapa[novaPosY][novaPosX] == portal {
			destX, destY := 70, 25
			verificarPortal(destX, destY)
			return
		} else if mapa[novaPosY][novaPosX] == vitoria {
			if moedasColetadas < totalDeMoedas {
				exibirAvisoMoedas()
				mapa[posY][posX] = ultimoElementoSobPersonagem
				posX, posY = posXInicial, posYInicial
				ultimoElementoSobPersonagem = vazio
				mapa[posY][posX] = personagem
			} else {
				animacaoVitoria(novaPosX, novaPosY)
				mensagemVitoria()
			}
			return
		} else if mapa[novaPosY][novaPosX] == moeda {
			if !moedasColetadasMap[novaPosY*100+novaPosX] {
				EfeitoMoeda(novaPosX, novaPosY)
				moedasColetadas++
				ultimoElementoSobPersonagem = vazio
			}
		}

		mapa[posY][posX] = ultimoElementoSobPersonagem

		ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX]
		posX, posY = novaPosX, novaPosY
		mapa[posY][posX] = personagem

		desenhaTudo()
	}
}

func moverVegetacao() {
	var cicloCount int
	const cicloMax = 10
	for {
		time.Sleep(1 * time.Second)

		mutex.Lock()
		for y, linha := range mapa {
			for x, elem := range linha {
				if elem.simbolo == vegetacao.simbolo || elem.simbolo == '♧' {

					if cicloCount < cicloMax {
						novaVariacao := vegetacaoVariacoes[rand.Intn(len(vegetacaoVariacoes))]
						mapa[y][x] = Elemento{simbolo: novaVariacao, cor: vegetacao.cor, corFundo: vegetacao.corFundo, tangivel: false}
					} else {

						mapa[y][x] = vegetacao
					}
				}
			}
		}
		mutex.Unlock()
		desenhaTudo()

		if cicloCount < cicloMax {
			cicloCount++
		} else {
			cicloCount = 0
		}
	}
}

func limparTela() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func exibeMensagem(msg string) {
	largura, altura := termbox.Size()
	x := (largura - len(msg)) / 2
	y := altura / 2
	for i, c := range msg {
		termbox.SetCell(x+i, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
}

func verificarMoedasEPosicionarEstrela() {
	if moedasColetadas == totalDeMoedas && !estrelaPosicionada {
		estrelaX, estrelaY := 76, 28
		mapa[estrelaY][estrelaX] = vitoria
		estrelaPosicionada = true
		desenhaTudo()
	}
}

func exibirAvisoMoedas() {
	mutex.Lock()
	defer mutex.Unlock()
	msg := fmt.Sprintf("Você precisa coletar %d moedas para vencer! Voltando para a base...", totalDeMoedas)
	largura, altura := termbox.Size()
	y := len(mapa) + 5

	if y >= altura {
		y = altura - 1
	}

	for i := 0; i < largura; i++ {
		termbox.SetCell(i, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
	}

	for i, c := range msg {
		termbox.SetCell(i, y, c, termbox.ColorRed, termbox.ColorDefault)
	}

	termbox.Flush()

	time.Sleep(time.Second * 2)

	for i := 0; i < largura; i++ {
		termbox.SetCell(i, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.Flush()
}

func mensagemVitoria() {
	mutex.Lock()
	defer mutex.Unlock()
	msgVitoria := "Parabens! Voce ganhou! Pressione ESC para sair."

	termbox.Flush()
	y := len(mapa) + 5

	for {
		for i, c := range msgVitoria {
			termbox.SetCell(i, y, c, termbox.ColorGreen, termbox.ColorDefault)
		}
		termbox.Flush()
		time.Sleep(time.Second * 2)
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				termbox.Close()
				os.Exit(0)
			}
		}
	}
}

func animacaoVitoria(x, y int) {
	jogoEmAndamento = false
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			termbox.SetCell(x, y, vitoria.simbolo, termbox.ColorYellow, termbox.ColorDefault)
		} else {
			termbox.SetCell(x, y, vitoria.simbolo, termbox.ColorLightMagenta, termbox.ColorDefault)
		}
		termbox.Flush()
		time.Sleep(100 * time.Millisecond)
	}

}

func verificarPortal(destX, destY int) {
	efeitoPortal(posX, posY)
	mapa[posY][posX] = vazio

	posX, posY = destX, destY
	ultimoElementoSobPersonagem = mapa[posY][posX]
	mapa[posY][posX] = personagem
	desenhaTudo()
}

func efeitoPortal(posX, posY int) {
	originalChar := mapa[posY][posX]
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			mapa[posY][posX] = portal
		} else {
			mapa[posY][posX] = originalChar
		}
		desenhaTudo()
		time.Sleep(100 * time.Millisecond)
	}

}

func EfeitoMoeda(x, y int) {
	if !moedasColetadasMap[y*100+x] {
		termbox.SetCell(x, y, moeda.simbolo, termbox.ColorYellow, termbox.ColorYellow)
		termbox.Flush()
		time.Sleep(200 * time.Millisecond)
		termbox.SetCell(x, y, vazio.simbolo, termbox.ColorDefault, termbox.ColorDefault)
		termbox.Flush()
		moedasColetadasMap[y*100+x] = true
	}

	verificarMoedasEPosicionarEstrela()
}

func EfeitoDano(x, y int) {

	termbox.SetCell(x, y, inimigo.simbolo, termbox.ColorDefault, termbox.ColorRed)
	termbox.Flush()
	time.Sleep(200 * time.Millisecond)

	if mapa[y][x] == inimigo {
		termbox.SetCell(x, y, inimigo.simbolo, termbox.ColorDefault, termbox.ColorDefault)
	} else {
		termbox.SetCell(x, y, vazio.simbolo, termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.Flush()
}

func GameOver() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	msgGameOver := "Game Over! Pressione ESC para sair."
	largura, altura := termbox.Size()
	for i, c := range msgGameOver {
		x := (largura-len(msgGameOver))/2 + i
		y := altura / 2
		termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				termbox.Close()
				os.Exit(0)
			}
		}
	}
}

func StartGame() {
	startTime = time.Now()
	jogoEmAndamento = true
	piscarMensagem = true
	go piscarMensagemInicio()

	termbox.PollEvent()

	piscarMensagem = false
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
}

func piscarMensagemInicio() {
	msg := "Pressione qualquer tecla para comecar"
	largura, altura := termbox.Size()

	cores := []termbox.Attribute{termbox.ColorRed, termbox.ColorGreen, termbox.ColorYellow, termbox.ColorBlue, termbox.ColorMagenta, termbox.ColorCyan}
	corIndex := 0

	for piscarMensagem {

		for i, c := range msg {
			x := (largura-len(msg))/2 + i
			y := altura / 2
			termbox.SetCell(x, y, c, cores[corIndex], termbox.ColorDefault)
		}
		termbox.Flush()

		time.Sleep(500 * time.Millisecond)

		corIndex++
		if corIndex >= len(cores) {
			corIndex = 0
		}
	}

	for i := range msg {
		x := (largura-len(msg))/2 + i
		y := altura / 2
		termbox.SetCell(x, y, vazio.simbolo, termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.Flush()
}

func interagir() {
	statusMsg = fmt.Sprintf("Interagindo em (%d, %d)", posX, posY)
}
