# mrtn - *Minimal Runtime Test Node* 🎈

---

## Objetivo

O **`mrtn`** foi pensado para agilizar o ciclo de treino dos competidores da Maratona de Programação da SBC. Em vez de testar manualmente caso a caso ou depender de submissões em juízes online, o `mrtn`:

1. Sincroniza automaticamente os casos de teste e respostas oficiais a partir do repositório [`mrtn-oracle`](https://github.com/murilo-henrique060/mrtn-oracle) no GitHub.
2. Compila e executa localmente o seu código contra as entradas e saídas esperadas.
3. Exibe um relatório visual detalhado (Accepted, Wrong Answer, Time Limit Exceeded, Runtime Error).
4. É multiplataforma: compatível tanto com **Linux** quanto com **Windows**.

---

## Comandos

O `mrtn` mantém uma interface minimalista e focada em produtividade:

### 1. Sincronizar casos de teste (`mrtn sync`)

Baixa ou atualiza o repositório contendo as questões, casos de entrada e gabaritos esperados ([`mrtn-oracle`](https://github.com/murilo-henrique060/mrtn-oracle)):

```bash
mrtn sync
```

> **Nota:** Os casos de teste sincronizados são obtidos de [murilo-henrique060/mrtn-oracle](https://github.com/murilo-henrique060/mrtn-oracle) e salvos no diretório de dados do usuário (`~/.local/share/mrtn/oracle` no Linux, `%LocalAppData%\mrtn\oracle` no Windows e `~/Library/Application Support/mrtn/oracle` no macOS, podendo ser alterado via `MRTN_DATA_DIR`), permitindo treinar e rodar testes mesmo sem conexão com a internet.

---

### 2. Consultar informações (`mrtn info`)

Permite inspecionar as provas, questões e casos de teste disponíveis no cache local:

```bash
mrtn info [prova] [questao] [numero-do-teste]
```

O comando adapta as informações exibidas de acordo com os parâmetros fornecidos:

* **Sem argumentos (`mrtn info`):**  
  Lista todas as edições/provas sincronizadas no sistema (ex: `2026-1`, referente à 1ª fase de 2026).
  ```bash
  mrtn info
  ```

* **Informando a prova (`mrtn info <prova>`):**  
  Lista todas as questões daquela edição com suas letras identificadoras e títulos oficiais (extraídos do `problem.yaml`).
  ```bash
  mrtn info 2026-1
  ```

* **Informando prova e questão (`mrtn info <prova> <questao>`):**  
  Lista todos os casos de teste disponíveis para a questão (amostras e testes fechados), numerados sequencialmente a partir de `1`.
  ```bash
  mrtn info 2026-1 A
  ```

* **Informando prova, questão e número do teste (`mrtn info <prova> <questao> <numero-do-teste>`):**  
  Exibe informações detalhadas sobre o caso de teste específico, incluindo o conteúdo de entrada (`input`), saída esperada (`output`) e metadados.  
  > **Importante:** Apenas o número do teste (na ordem exibida pelo `mrtn info`, começando em `1`) é válido — por exemplo, use `1` ao invés de `sample001`.
  ```bash
  mrtn info 2026-1 A 1
  ```

---

### 3. Executar testes (`mrtn test`)

Executa o seu código contra os casos de teste da questão especificada, informando a prova e a questão (*case-insensitive*, ex: `2026-1 a` ou `2026-1 A`):

```bash
mrtn test <prova> <questao> [numero-do-teste] <arquivo>
```

> **Nota sobre `[numero-do-teste]`:** Assim como no `mrtn info`, apenas o índice numérico do teste (começando em `1`, seguindo a ordem listada no `info`) é aceito pelo programa (ex: use `1` ao invés de `sample001`).

#### Exemplos de uso:

* **Testar todos os casos da questão `A` da prova `2026-1`:**
  ```bash
  mrtn test 2026-1 A solucao.cpp
  ```

* **Testar apenas um caso de teste específico (ex: caso `3` da questão `B` de `2026-1`, aceitando minúsculas):**
  ```bash
  mrtn test 2026-1 b 3 solucao.py
  ```

* **Testar solução em Python para a questão `M` de `2026-1`:**
  ```bash
  mrtn test 2026-1 m solucao.py
  ```

---

### 4. Limpar dados locais (`mrtn clear`)

Remove os dados das provas, questões e casos de teste sincronizados no diretório de dados do usuário:

```bash
mrtn clear
```

> **Nota:** Após a limpeza, você pode baixar os casos de teste novamente a qualquer momento executando `mrtn sync`.

---

## Suporte a Linguagens

As linguagens oficiais da Maratona de Programação SBC são **C, C++, Java, Kotlin e Python**.

* **Versão atual:** Suporte completo para todas as 5 linguagens oficiais: **Python (`.py`)**, **C (`.c`)**, **C++ (`.cpp`, `.cc`, `.cxx`)**, **Java (`.java`)** e **Kotlin (`.kt`)**.

| Linguagem | Extensão | Status Atual | Compilador / Interpretador |
| :--- | :--- | :--- | :--- |
| **Python** | `.py` | ✅ Suportado | `pypy3` / `python3` / `python` (`py_compile`) |
| **C** | `.c` | ✅ Suportado | `gcc -std=gnu11 -O2 -lm -static` |
| **C++** | `.cpp`, `.cc`, `.cxx` | ✅ Suportado | `g++ -std=c++20 -O2 -lm -static` |
| **Java** | `.java` | ✅ Suportado | `javac` + `jar` & `java` (`-Xmx`, `-Xss131072k`) |
| **Kotlin** | `.kt` | ✅ Suportado | `kotlinc -include-runtime` & `java` (`-Xmx`, `-Xss131072k`) |

---

## Instalação e Requisitos

### Pré-requisitos
* [Git](https://git-scm.com/) instalado no sistema (para clonagem/sincronização dos casos de teste com o `mrtn sync`).
* Compiladores e interpretadores das linguagens que você pretende utilizar (ex: `pypy3` / `python3`, `gcc`, `g++`, `javac` / `java`, `kotlinc`).

---

### 1. Instalação Automática (Recomendado)

Instale o `mrtn` com apenas um comando no terminal (não necessita ter o Go instalado):

#### Linux & macOS (Terminal)
```bash
curl -fsSL https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.sh | bash
```
> O instalador detecta automaticamente seu sistema e arquitetura (`amd64` ou `arm64`), instala o binário em `~/.local/bin` (ou `/usr/local/bin`) e orienta a configuração do `PATH`.

#### Windows (PowerShell)
Abra o **PowerShell** e execute:
```powershell
irm https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.ps1 | iex
```
> O script baixa o `mrtn.exe` compatível, coloca em `%LOCALAPPDATA%\Programs\mrtn` e adiciona automaticamente à variável de ambiente `PATH` do usuário.

#### Windows (Duplo Clique via Arquivo `.bat`)
Se preferir não digitar comandos:
1. Baixe os arquivos [`install.bat`](https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.bat) e [`install.ps1`](https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.ps1) na mesma pasta.
2. Dê **duplo clique** em `install.bat`.

---

### 2. Via `go install` (Para desenvolvedores Go)

```bash
go install github.com/murilo-henrique060/mrtn-cli/cmd/mrtn@latest
```

> **Nota:** Certifique-se de que o diretório de binários do Go (`$GOPATH/bin` ou `~/go/bin`) esteja no seu `PATH`:
> - **Linux/macOS:** `export PATH=$PATH:$(go env GOPATH)/bin` (adicione ao seu `~/.bashrc` ou `~/.zshrc`)
> - **Windows:** Verifique se `%USERPROFILE%\go\bin` está presente na variável de ambiente `PATH`.

---

### 3. Via binários pré-compilados manuais

Baixe o executável correspondente ao seu sistema operacional na página de [Releases](https://github.com/murilo-henrique060/mrtn-cli/releases/tag/v1.0.1):

| Sistema | Arquitetura | Pacote (`v1.0.1`) |
| :--- | :--- | :--- |
| **Linux** | x86_64 (`amd64`) | `mrtn-v1.0.1-linux-amd64` |
| **Linux** | ARM64 (`arm64`) | `mrtn-v1.0.1-linux-arm64` |
| **Windows** | x86_64 (`amd64`) | `mrtn-v1.0.1-windows-amd64.exe` |
| **Windows** | ARM64 (`arm64`) | `mrtn-v1.0.1-windows-arm64.exe` |
| **macOS** | Intel (`amd64`) | `mrtn-v1.0.1-darwin-amd64` |
| **macOS** | Apple Silicon (`arm64`) | `mrtn-v1.0.1-darwin-arm64` |

#### Linux
```bash
curl -LO https://github.com/murilo-henrique060/mrtn-cli/releases/download/v1.0.1/mrtn-v1.0.1-linux-amd64
chmod +x mrtn-v1.0.1-linux-amd64
sudo mv mrtn-v1.0.1-linux-amd64 /usr/local/bin/mrtn
mrtn --version
```

#### macOS
```bash
curl -LO https://github.com/murilo-henrique060/mrtn-cli/releases/download/v1.0.1/mrtn-v1.0.1-darwin-arm64
chmod +x mrtn-v1.0.1-darwin-arm64
sudo mv mrtn-v1.0.1-darwin-arm64 /usr/local/bin/mrtn
mrtn --version
```

#### Windows (Manual via PowerShell)
```powershell
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\Programs\mrtn" | Out-Null
Invoke-WebRequest `
  -Uri "https://github.com/murilo-henrique060/mrtn-cli/releases/download/v1.0.1/mrtn-v1.0.1-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\mrtn\mrtn.exe"

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$env:LOCALAPPDATA\Programs\mrtn*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:LOCALAPPDATA\Programs\mrtn", "User")
}
```

> **Verificação de integridade (opcional):** Os hashes SHA-256 dos binários estão disponíveis em `checksums.txt` na release (`sha256sum -c checksums.txt --ignore-missing`).

---

### 4. Compilando a partir do código-fonte

Para desenvolvimento ou compilação manual:

```bash
# Clone o repositório
git clone https://github.com/murilo-henrique060/mrtn-cli.git
cd mrtn-cli

# Instale o binário diretamente no seu GOPATH/bin
go install ./cmd/mrtn

# Ou apenas compile o executável no diretório atual
go build -o mrtn ./cmd/mrtn
```

---

## Roadmap de Desenvolvimento

- [x] Definição do escopo e comandos principais
- [x] Inicialização do módulo Go (`go mod init`)
- [x] Estrutura base da CLI (`mrtn sync`, `mrtn info`, `mrtn test` e `mrtn clear`)
- [x] Implementação de `mrtn sync` (integração Git / GitHub com `mrtn-oracle`)
- [x] Implementação de `mrtn info` (inspeção de provas, questões e testes)
- [x] Implementação dos runners para Python (`.py`), C (`.c`), C++ (`.cpp`, `.cc`, `.cxx`), Java (`.java`) e Kotlin (`.kt`)
- [x] Comparador de saídas com diff formatado e normalização de quebras de linha
- [x] Implementação de timeouts (TLE) e tratamento de erros de runtime (RTE)
- [x] Suporte multiplataforma testado em Linux e Windows

---

## Licença

Distribuído sob a licença MIT. Consulte `LICENSE` para mais informações.
