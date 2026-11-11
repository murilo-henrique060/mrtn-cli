# mrtn 🎈

> **mrtn** (*Minimal Runtime Test Node* ou *Meu Resultado Terminou Negado*)  
> Utilitário CLI em Go para treino local da Maratona SBC de Programação (ICPC Brasil).

---

## 🎯 Objetivo

O **`mrtn`** foi pensado para agilizar o ciclo de treino dos competidores da Maratona de Programação da SBC. Em vez de testar manualmente caso a caso ou depender de submissões em juízes online com filas demoradas, o `mrtn`:

1. Sincroniza automaticamente os casos de teste e respostas oficiais a partir de um repositório no GitHub.
2. Compila e executa localmente o seu código contra as entradas e saídas esperadas.
3. Exibe um relatório visual detalhado (Accepted, Wrong Answer, Time Limit Exceeded, Runtime Error).
4. É multiplataforma: compatível tanto com **Linux** quanto com **Windows**.

---

## 🚀 Comandos

O `mrtn` mantém uma interface minimalista e focada em produtividade:

### 1. Sincronizar casos de teste (`mrtn sync`)

Baixa ou atualiza o repositório contendo as questões, casos de entrada (`.in`) e gabaritos esperados (`.sol` / `.out`):

```bash
mrtn sync
```

> **Nota:** Os casos de teste sincronizados ficam armazenados em cache local, permitindo treinar e rodar testes mesmo sem conexão com a internet.

---

### 2. Executar testes (`mrtn test`)

Executa o seu código contra os casos de teste da questão especificada:

```bash
mrtn test <questao> [numero-do-teste] <arquivo>
```

#### Exemplos de uso:

* **Testar todos os casos da questão `A`:**
  ```bash
  mrtn test A solucao.cpp
  ```

* **Testar apenas um caso de teste específico (ex: caso `3` da questão `B`):**
  ```bash
  mrtn test B 3 solucao.py
  ```

* **Testar solução em Go para a questão `M`:**
  ```bash
  mrtn test M main.go
  ```

---

## 🛠️ Suporte a Linguagens

O runner do `mrtn` detecta automaticamente a linguagem pelo formato do arquivo e aplica as flags recomendadas para maratona:

| Linguagem | Extensão | Compilador / Interpretador |
| :--- | :--- | :--- |
| **C++** | `.cpp`, `.cc` | `g++ -O2 -std=c++17` |
| **C** | `.c` | `gcc -O2 -std=c11` |
| **Python** | `.py` | `python3` / `pypy3` |
| **Java** | `.java` | `javac` & `java` |
| **Go** | `.go` | `go run` / `go build` |

---

## 📦 Instalação e Requisitos

### Pré-requisitos
* [Git](https://git-scm.com/) instalado no sistema (para clonagem/sincronização dos casos de teste).
* [Go](https://go.dev/) (versão 1.21 ou superior) para compilação do binário.
* Compiladores e interpretadores das linguagens que você pretende utilizar (ex: `g++`, `python3`).

### Compilando a partir do código-fonte

```bash
# Clone o repositório
git clone https://github.com/<seu-usuario>/mrtn.git
cd mrtn

# Compile o binário
go build -o mrtn .

# (Opcional) Mova para o PATH do seu sistema
# No Linux:
sudo mv mrtn /usr/local/bin/

# No Windows:
# Adicione a pasta do executável mrtn.exe às Variáveis de Ambiente do Sistema (PATH)
```

---

## 🗺️ Roadmap de Desenvolvimento

- [x] Definição do escopo e comandos principais
- [ ] Inicialização do módulo Go (`go mod init`)
- [ ] Estrutura base da CLI (Cobra ou flags nativas)
- [ ] Implementação de `mrtn sync` (integração Git / GitHub)
- [ ] Implementação do runner e compilação das linguagens suportadas
- [ ] Comparador de saídas com diff formatado e normalização de quebras de linha
- [ ] Implementação de timeouts (TLE) e tratamento de erros de runtime (RTE)
- [ ] Suporte multiplataforma testado em Linux e Windows

---

## 📄 Licença

Distribuído sob a licença MIT. Consulte `LICENSE` para mais informações.
