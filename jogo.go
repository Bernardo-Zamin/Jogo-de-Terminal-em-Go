package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

var personagem rune = '☺'
var parede rune = '▣'
var vegetacao rune = '♣'
var inimigo rune = '▶'
var moeda rune = '◉'
var totalDeMoedas int = 10
var portal rune = '⚑'
var vitoria rune = '★'
var estrelaPosicionada bool
var piscarMensagem bool

var mapa [][]rune
var posX, posY int
var ultimoCharSobPersonagem rune = ' '
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
	mapa[posY][posX] = personagem
	desenhaTudo()
	go moverInimigo()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return
			}
			comando := rune(0)

			if ev.Ch == 'w' {
				comando = 'w'
			} else if ev.Ch == 'a' {
				comando = 'a'
			} else if ev.Ch == 's' {
				comando = 's'
			} else if ev.Ch == 'd' {
				comando = 'd'
			}
			mover(comando)
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
		var linhaRune []rune
		linha := scanner.Text()
		for x, char := range linha {
			if char == personagem {
				posX, posY = x, y
				char = ' '
			}
			linhaRune = append(linhaRune, char)
		}
		mapa = append(mapa, linhaRune)
		y++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	mapa[posY][posX] = personagem
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, linha := range mapa {
		for x, char := range linha {
			var cor termbox.Attribute
			switch char {
			case parede:
				cor = termbox.ColorDefault | termbox.AttrBold | termbox.AttrUnderline
			case vegetacao:
				cor = termbox.ColorGreen | termbox.AttrBold
			case inimigo:
				cor = termbox.ColorRed | termbox.AttrBold
			case moeda:
				cor = termbox.ColorYellow | termbox.AttrBold
			case personagem:
				cor = termbox.ColorBlue | termbox.AttrBold
			case portal:
				cor = termbox.ColorCyan | termbox.AttrBold
			case vitoria:
				cor = termbox.ColorLightMagenta | termbox.AttrBold
			default:
				cor = termbox.ColorDefault
			}
			termbox.SetCell(x, y, char, cor, termbox.ColorDefault)
		}
	}

	msg := fmt.Sprintf("Use WASD para mover. Pressione ESC para sair. Moedas: %d Vidas: %d", moedasColetadas, vidas)
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+2, c, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
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

				termbox.SetCell(posX, posY, personagem, termbox.ColorDefault, termbox.ColorRed)
				termbox.Flush()
				time.Sleep(200 * time.Millisecond)

				if posX > 0 && mapa[posY][posX-1] == ' ' {
					mapa[posY][posX] = ' '
					posX--
					ultimoCharSobPersonagem = ' '
					mapa[posY][posX] = personagem
				}

				if vidas == 0 {
					GameOver()
					mutex.Unlock()
					return
				}
			}

			mapa[inimigoY][inimigoX] = ' '
			mapa[novaY][inimigoX] = inimigo
			inimigoY = novaY
		}
		mutex.Unlock()
		desenhaTudo()
	}
}

func mover(comando rune) {
	mutex.Lock()
	defer mutex.Unlock()

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

	// Impede que o personagem se mova para uma posição ilegal
	if novaPosY < 0 || novaPosY >= len(mapa) || novaPosX < 0 || novaPosX >= len(mapa[novaPosY]) ||
		mapa[novaPosY][novaPosX] == parede {
		return
	}

	// Se o personagem tentar se mover para o portal, verifica e executa a ação necessária
	if mapa[novaPosY][novaPosX] == portal {
		verificarPortal(70, 25) // Substitua por suas coordenadas específicas
		desenhaTudo()
		return
	}

	// Se o personagem tentar se mover para a posição da vitória, verifica se todas as moedas foram coletadas
	if mapa[novaPosY][novaPosX] == vitoria {
		if moedasColetadas < totalDeMoedas {
			exibirAvisoMoedas()
			return // Não permite que o personagem alcance a vitória sem coletar todas as moedas
		} else {
			animacaoVitoria(posX, posY) // Executa a animação de vitória
			return
		}
	}

	// Se o personagem tentar se mover para uma posição com um inimigo
	if mapa[novaPosY][novaPosX] == inimigo {
		vidas--
		EfeitoDano(novaPosX, novaPosY)
		if vidas == 0 {
			GameOver()
			return
		}
	} else {
		// Se a nova posição contiver uma moeda, incrementa o contador de moedas coletadas
		if mapa[novaPosY][novaPosX] == moeda {
			moedasColetadas++
			mapa[novaPosY][novaPosX] = ' '
			EfeitoMoeda(novaPosX, novaPosY)
		}

		if mapa[novaPosY][novaPosX] == vegetacao {
			mapa[novaPosY][novaPosX] = ' '
			mapa[posY][posX] = vegetacao

		}

		// Atualiza a posição do personagem
		ultimoCharSobPersonagem = mapa[novaPosY][novaPosX]
		mapa[posY][posX] = ultimoCharSobPersonagem
		posX, posY = novaPosX, novaPosY
		mapa[posY][posX] = personagem
	}

	desenhaTudo()
}

func limparCentroDaTela() {
	largura, altura := termbox.Size()
	for y := altura/2 - 1; y <= altura/2+1; y++ {
		for x := 0; x < largura; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func exibirAvisoMoedas() {
	msg := "Voce precisa coletar todas as 10 moedas para ganhar!"
	largura, altura := termbox.Size()
	x := (largura - len(msg)) / 2
	y := altura / 2
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault) // Limpa a tela antes de exibir a mensagem
	for i, c := range msg {
		termbox.SetCell(x+i, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
	time.Sleep(3 * time.Second) // Exibe a mensagem por 3 segundos

	// Limpa a posição atual do personagem antes de movê-lo de volta para a posição inicial.
	mapa[posY][posX] = ultimoCharSobPersonagem
	posX, posY = 8, 8             // Coordenadas da posição inicial, ajuste conforme necessário
	ultimoCharSobPersonagem = ' ' // Assumindo que a posição inicial está vazia
	mapa[posY][posX] = personagem
	desenhaTudo()
}

func exibirMensagem(msg string) {
	largura, altura := termbox.Size()
	x := (largura - len(msg)) / 2
	y := altura / 2
	limparCentroDaTela()
	for i, c := range msg {
		termbox.SetCell(x+i, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
	time.Sleep(3 * time.Second) // Exibe a mensagem por 3 segundos
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

func verificarPortal(destX, destY int) {
	go efeitoPortal(posX, posY)
	mapa[posY][posX] = ' '
	posX, posY = destX, destY
	ultimoCharSobPersonagem = mapa[posY][posX]
	mapa[posY][posX] = personagem
}

func animacaoVitoria(x, y int) {
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			termbox.SetCell(x, y, vitoria, termbox.ColorYellow, termbox.ColorDefault)
		} else {
			termbox.SetCell(x, y, vitoria, termbox.ColorLightMagenta, termbox.ColorDefault)
		}
		termbox.Flush()
		time.Sleep(200 * time.Millisecond)
	}
	celebreVitoria()
}

func efeitoPortal(posX, posY int) {
	originalChar := mapa[posY][posX]
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			mapa[posY][posX] = '⚑'
		} else {
			mapa[posY][posX] = originalChar
		}
		desenhaTudo()
		time.Sleep(100 * time.Millisecond)
	}
	mapa[posY][posX] = originalChar
	desenhaTudo()
}

func EfeitoMoeda(x, y int) {
	termbox.SetCell(x, y, moeda, termbox.ColorYellow, termbox.ColorYellow)
	termbox.Flush()
	time.Sleep(200 * time.Millisecond)
	termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()

	verificarMoedasEPosicionarEstrela()

}

func EfeitoDano(x, y int) {

	termbox.SetCell(x, y, inimigo, termbox.ColorDefault, termbox.ColorRed)
	termbox.Flush()
	time.Sleep(200 * time.Millisecond)

	if mapa[y][x] == inimigo {
		termbox.SetCell(x, y, inimigo, termbox.ColorDefault, termbox.ColorDefault)
	} else {
		termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
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
		termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
	termbox.Flush()
}
