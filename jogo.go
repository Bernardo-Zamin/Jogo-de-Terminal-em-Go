package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// Define os elementos do jogo
type Elemento struct {
	simbolo  rune
	cor      termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}

// Personagem controlado pelo jogador
var personagem = Elemento{
	simbolo:  '☺',
	cor:      termbox.ColorWhite,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var parede = Elemento{
	simbolo:  '▣',
	cor:      termbox.ColorRed,
	corFundo: termbox.ColorDefault,
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
	cor:      termbox.ColorBlue,
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

// Elemento vazio
var vazio = Elemento{
	simbolo:  ' ',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
	simbolo:  '.',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorYellow,
	tangivel: false,
}

var estrelaPosicionada bool
var piscarMensagem bool

var moedasColetadasMap = make(map[int]bool) // Mapa para rastrear quais moedas foram coletadas

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

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return // Sair do programa
			}
			if ev.Ch == 'e' {
				interagir()
			} else {
				mover(ev.Ch)
			}
			desenhaTudo()
		}
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
				posXInicial, posYInicial = x, y // Armazena a posição inicial
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
					elem = vazio // Se a moeda foi coletada, não desenhe ela
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
	for i, c := range statusMsg {
		termbox.SetCell(i, len(mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair. Moedas coletadas: " + fmt.Sprintf("%d", moedasColetadas) + " Vidas: " + fmt.Sprintf("%d", vidas)
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}
}

func revelarArea() {
	minX := max(0, posX-raioVisao)
	maxX := min(len(mapa[0])-1, posX+raioVisao)
	minY := max(0, posY-raioVisao/2)
	maxY := min(len(mapa)-1, posY+raioVisao/2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Revela as células dentro do quadrado de visão
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
		// Se o novo local é tangível ou é um portal, pare a movimentação.
		if mapa[novaPosY][novaPosX].tangivel && mapa[novaPosY][novaPosX] != portal {
			return
		}

		// Verifica se a nova posição contém um portal
		if mapa[novaPosY][novaPosX] == portal {
			destX, destY := 70, 25 // Defina as coordenadas de destino do portal
			verificarPortal(destX, destY)
			return // Pare a função aqui para evitar redesenhar o personagem na posição antiga
		} else if mapa[novaPosY][novaPosX] == vitoria {
			if moedasColetadas < totalDeMoedas {
				exibirAvisoMoedas()
			} else {
				celebreVitoria()
			}
			return // Encerrar a função para não mover o personagem após o jogo ter acabado ou mostrar a mensagem
		} else if mapa[novaPosY][novaPosX] == moeda {
			if !moedasColetadasMap[novaPosY*100+novaPosX] { // Verifica se a moeda já foi coletada
				EfeitoMoeda(novaPosX, novaPosY)
				moedasColetadas++
				ultimoElementoSobPersonagem = vazio // Remove a moeda da posição após coletada
			}
		}

		// Atualiza o local onde o personagem estava anteriormente
		mapa[posY][posX] = ultimoElementoSobPersonagem
		// Move o personagem para a nova posição
		ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX]
		posX, posY = novaPosX, novaPosY
		mapa[posY][posX] = personagem

		desenhaTudo()
	}
}

func limparCentroDaTela() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	// certifique-se de chamar termbox.Flush() se for necessário aplicar imediatamente a limpeza
}

func exibirAvisoMoedas() {
	if moedasColetadas < totalDeMoedas {
		// Mensagem de aviso
		msg := fmt.Sprintf("Você precisa coletar todas as %d moedas para ganhar!", totalDeMoedas)

		// Limpa a tela central onde a mensagem será exibida
		limparCentroDaTela()

		// Exibe a mensagem
		exibirMensagem(msg)

		// Atualiza a tela para exibir a mensagem
		termbox.Flush()

		// Espera por 3 segundos antes de continuar, sem qualquer outra alteração visual ou lógica
		time.Sleep(5000 * time.Millisecond)

		// Reseta o personagem para a posição inicial após a pausa
		mapa[posY][posX] = ultimoElementoSobPersonagem
		posX, posY = posXInicial, posYInicial // Usando as coordenadas de posição inicial armazenadas
		ultimoElementoSobPersonagem = vazio   // Assumindo que a posição inicial está vazia
		mapa[posY][posX] = personagem

		// Não precisa chamar Flush aqui, pois a próxima ação de desenho no loop principal já o fará
	}
}

func exibirMensagem(msg string) {
	largura, altura := termbox.Size()
	x := (largura - len(msg)) / 2
	y := altura / 2
	for i, c := range msg {
		termbox.SetCell(x+i, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	// termbox.Flush() chamado após a configuração de todas as células para sincronizar a renderização
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

func celebreVitoria() {
	go piscarMensagemVitoria()

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

func piscarMensagemVitoria() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	msgVitoria := "Parabens! Você ganhou! Pressione ESC para sair."
	largura, altura := termbox.Size()
	termbox.Flush()
	for {

		for i, c := range msgVitoria {
			x := (largura-len(msgVitoria))/2 + i
			y := altura / 2
			termbox.SetCell(x, y, c, termbox.ColorGreen, termbox.ColorDefault)
		}
		termbox.Flush()
		time.Sleep(500 * time.Millisecond)

		for i, c := range msgVitoria {
			x := (largura-len(msgVitoria))/2 + i
			y := altura / 2
			termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
		}
		termbox.Flush()
		time.Sleep(500 * time.Millisecond)
	}
}
func animacaoVitoria(x, y int) {
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			termbox.SetCell(x, y, vitoria.simbolo, termbox.ColorYellow, termbox.ColorDefault)
		} else {
			termbox.SetCell(x, y, vitoria.simbolo, termbox.ColorLightMagenta, termbox.ColorDefault)
		}
		termbox.Flush()
		time.Sleep(200 * time.Millisecond)
	}
	celebreVitoria()
}

func verificarPortal(destX, destY int) {
	efeitoPortal(posX, posY) // Execute a animação do portal e espere até que termine
	mapa[posY][posX] = vazio // Limpe a posição antiga do personagem
	// Atualize a posição do personagem
	posX, posY = destX, destY
	ultimoElementoSobPersonagem = mapa[posY][posX] // Prepare o estado para a nova posição
	mapa[posY][posX] = personagem                  // Coloque o personagem na nova posição
	desenhaTudo()                                  // Redesenhe o jogo para refletir a nova posição do personagem
}

// A função efeitoPortal não precisa da goroutine agora
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
	// Não altere o mapa após a animação, pois isso será tratado em verificarPortal
}

func EfeitoMoeda(x, y int) {
	if !moedasColetadasMap[y*100+x] { // Verifique se a moeda já foi coletada
		termbox.SetCell(x, y, moeda.simbolo, termbox.ColorYellow, termbox.ColorYellow)
		termbox.Flush()
		time.Sleep(200 * time.Millisecond)
		termbox.SetCell(x, y, vazio.simbolo, termbox.ColorDefault, termbox.ColorDefault)
		termbox.Flush()
		moedasColetadasMap[y*100+x] = true // Marque a moeda como coletada
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
	piscarMensagem = true
	go piscarMensagemInicio() // Inicia a goroutine para piscar a mensagem de início

	termbox.PollEvent() // Aguarda o usuário pressionar qualquer tecla para começar

	piscarMensagem = false
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault) // Limpa a tela para começar o jogo
	termbox.Flush()
}

func piscarMensagemInicio() {
	msg := "Pressione qualquer tecla para comecar"
	largura, altura := termbox.Size()

	cores := []termbox.Attribute{termbox.ColorRed, termbox.ColorGreen, termbox.ColorYellow, termbox.ColorBlue, termbox.ColorMagenta, termbox.ColorCyan}
	corIndex := 0

	for piscarMensagem { // Continua piscando até que piscarMensagem seja falso
		// Pinta a mensagem com a cor atual
		for i, c := range msg {
			x := (largura-len(msg))/2 + i
			y := altura / 2
			termbox.SetCell(x, y, c, cores[corIndex], termbox.ColorDefault)
		}
		termbox.Flush()

		// Espera um pouco antes de mudar a cor
		time.Sleep(500 * time.Millisecond)

		// Muda para a próxima cor
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
