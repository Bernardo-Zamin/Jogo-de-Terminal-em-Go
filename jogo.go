package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

var personagem rune = '☺' // Personagem
var parede rune = '▣'     // Parede
var barreira rune = '#'   // Barreira
var vegetacao rune = '♣'  // Vegetação
var inimigo rune = '▶'    // Inimigo
var moeda rune = '◉'      // Moeda

var mapa [][]rune                      // Alterado para [][]rune
var posX, posY int                     // Posição inicial do personagem
var ultimoCharSobPersonagem rune = ' ' // Assume que o chão inicial é vazio
var moedasColetadas int                // Contagem de moedas coletadas
var vidas = 5                          // Vidas do personagem
var mutex sync.Mutex

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	carregarMapa("mapa.txt")
	mapa[posY][posX] = personagem
	desenhaTudo()
	go moverInimigo() // Inicia a goroutine para mover o inimigo

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return // Sair do programa
			}
			comando := rune(0) // Inicializa como rune nulo

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
	y := 0 // Inicializa o contador de linhas
	for scanner.Scan() {
		var linhaRune []rune
		linha := scanner.Text()
		for x, char := range linha {
			if char == personagem {
				posX, posY = x, y // Salva a posição inicial do personagem
				char = ' '        // Substitui o personagem por um espaço no mapa
			}
			linhaRune = append(linhaRune, char)
		}
		mapa = append(mapa, linhaRune)
		y++ // Incrementa o contador de linhas após processar cada linha
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// Agora que o mapa foi carregado, coloque o personagem na sua posição inicial.
	mapa[posY][posX] = personagem // Isso garante que haja apenas um personagem no mapa.
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, linha := range mapa {
		for x, char := range linha {
			var cor termbox.Attribute
			switch char {
			case parede:
				cor = termbox.ColorBlue
			case barreira:
				cor = termbox.ColorMagenta
			case vegetacao:
				cor = termbox.ColorGreen
			case inimigo:
				cor = termbox.ColorRed
			case moeda:
				cor = termbox.ColorYellow
			case personagem:
				cor = termbox.ColorWhite
			default:
				cor = termbox.ColorDefault
			}
			termbox.SetCell(x, y, char, cor, termbox.ColorDefault)
		}
	}

	// Desenha a interface do usuário, como o contador de moedas e vidas
	msg := fmt.Sprintf("Use WASD para mover. Pressione ESC para sair. Moedas: %d Vidas: %d", moedasColetadas, vidas)
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+2, c, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
}

func moverInimigo() {
	var direcao int = 1 // 1 para "cima", -1 para "baixo"
	var movimentos int = 0
	const maxMovimentos = 2

	for {
		time.Sleep(500 * time.Millisecond) // Intervalo de movimento

		mutex.Lock()
		var inimigoX, inimigoY int
		encontrado := false
		for y, linha := range mapa {
			for x, char := range linha {
				if char == inimigo {
					inimigoX, inimigoY = x, y
					encontrado = true
					break // Sai do loop interno
				}
			}
			if encontrado {
				break // Sai do loop externo se o inimigo for encontrado
			}
		}

		if !encontrado {
			mutex.Unlock()
			continue // Se o inimigo não foi encontrado, pula para a próxima iteração do loop
		}

		novaY := inimigoY + direcao
		if movimentos >= maxMovimentos || novaY < 0 || novaY >= len(mapa) || mapa[novaY][inimigoX] == parede || mapa[novaY][inimigoX] == barreira {
			direcao *= -1
			movimentos = 0
		} else {
			movimentos++
			if novaY == posY && inimigoX == posX {
				vidas-- // O personagem perde uma vida

				// Efeito visual vermelho
				termbox.SetCell(posX, posY, personagem, termbox.ColorDefault, termbox.ColorRed)
				termbox.Flush()
				time.Sleep(200 * time.Millisecond) // Duração do efeito visual

				// Mover o personagem uma posição para trás se possível
				if posX > 0 && mapa[posY][posX-1] == ' ' {
					mapa[posY][posX] = ' ' // Limpa a posição atual do personagem
					posX--
					ultimoCharSobPersonagem = ' '
					mapa[posY][posX] = personagem // Redesenha o personagem na nova posição
				}

				if vidas == 0 {
					desenhaGameOver()
					mutex.Unlock()
					return // Termina a goroutine do inimigo
				}
			}

			// Muda o inimigo para a nova posição se não for parede ou barreira
			mapa[inimigoY][inimigoX] = ' '  // Limpa a posição antiga
			mapa[novaY][inimigoX] = inimigo // O inimigo ocupa a nova posição
			inimigoY = novaY                // Atualiza a posição Y do inimigo
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

	// Restringe o personagem de sair do mapa ou andar sobre paredes e barreiras
	if novaPosY < 0 || novaPosY >= len(mapa) || novaPosX < 0 || novaPosX >= len(mapa[novaPosY]) ||
		mapa[novaPosY][novaPosX] == parede || mapa[novaPosY][novaPosX] == barreira {
		return
	}

	if mapa[novaPosY][novaPosX] == inimigo {
		vidas-- // Personagem perde uma vida
		mostrarEfeitoDano(novaPosX, novaPosY)
		// Não mover o personagem para a célula do inimigo, manter ambos visíveis
		if vidas == 0 {
			desenhaGameOver()
			return
		}
	} else {
		// Se a célula destino é uma moeda, atualiza o contador e remove a moeda do mapa
		if mapa[novaPosY][novaPosX] == moeda {
			moedasColetadas++
			mapa[novaPosY][novaPosX] = ' ' // Remove a moeda do mapa
			mostrarEfeitoMoeda(novaPosX, novaPosY)
		}

		// Atualiza as posições do personagem e do que estava sob ele
		ultimoCharSobPersonagem = mapa[novaPosY][novaPosX]
		mapa[posY][posX] = ultimoCharSobPersonagem // Devolve o que estava sob o personagem
		posX, posY = novaPosX, novaPosY
		mapa[posY][posX] = personagem
	}

	desenhaTudo()
}

func mostrarEfeitoMoeda(x, y int) {
	termbox.SetCell(x, y, moeda, termbox.ColorYellow, termbox.ColorYellow)
	termbox.Flush()
	time.Sleep(200 * time.Millisecond)
	termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
}

func mostrarEfeitoDano(x, y int) {
	// Efeito visual de dano na célula do inimigo
	termbox.SetCell(x, y, inimigo, termbox.ColorDefault, termbox.ColorRed)
	termbox.Flush()
	time.Sleep(200 * time.Millisecond)

	// Se o inimigo ainda estiver lá, desenha-o novamente
	if mapa[y][x] == inimigo {
		termbox.SetCell(x, y, inimigo, termbox.ColorDefault, termbox.ColorDefault)
	} else {
		termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.Flush()
}

func desenhaGameOver() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault) // Limpa a tela

	msgGameOver := "Game Over! Pressione ESC para sair."
	largura, altura := termbox.Size() // Pega as dimensões do terminal
	for i, c := range msgGameOver {
		x := (largura-len(msgGameOver))/2 + i // Centraliza a mensagem
		y := altura / 2                       // Posiciona na metade da altura do terminal
		termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush() // Atualiza a tela

	// Aguarda o usuário pressionar ESC para sair
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				termbox.Close() // Fecha a interface gráfica
				os.Exit(0)      // Termina o programa
			}
		}
	}
}
