# mrtn - *Minimal Runtime Test Node* 🎈

O **`mrtn`** foi pensado para agilizar o ciclo de treino dos competidores da Maratona de Programação da SBC. Em vez de testar manualmente caso a caso ou depender de submissões em juízes online, o `mrtn`:

1. Sincroniza automaticamente os casos de teste e respostas oficiais a partir do repositório [`mrtn-oracle`](https://github.com/murilo-henrique060/mrtn-oracle) no GitHub.
2. Compila e executa localmente o seu código contra as entradas e saídas esperadas.
3. Exibe um relatório visual detalhado (Accepted, Wrong Answer, Time Limit Exceeded, Runtime Error).
4. É multiplataforma: compatível tanto com **Linux** quanto com **Windows**.

## Instalação e Requisitos

### Pré-requisitos
* [Git](https://git-scm.com/) instalado no sistema (para clonagem/sincronização dos casos de teste com o `mrtn sync`).
* Compiladores e interpretadores das linguagens que você pretende utilizar (ex: `pypy3` / `python3`, `gcc`, `g++`, `javac` / `java`, `kotlinc`).

### 1. Instalação Automática (Recomendado)

Instale o `mrtn` com apenas um comando no terminal (não necessita ter o Go instalado):

#### Linux & macOS (Terminal)
```bash
curl -fsSL https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.sh | bash
```

#### Windows (PowerShell)
Abra o **PowerShell** e execute:
```powershell
irm https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.ps1 | iex
```

### 2. Via `go install` (Para desenvolvedores Go)

```bash
go install github.com/murilo-henrique060/mrtn-cli/cmd/mrtn@latest
```

> **Nota:** Certifique-se de que o diretório de binários do Go (`$GOPATH/bin` ou `~/go/bin`) esteja no seu `PATH`:
> - **Linux/macOS:** `export PATH=$PATH:$(go env GOPATH)/bin` (adicione ao seu `~/.bashrc` ou `~/.zshrc`)
> - **Windows:** Verifique se `%USERPROFILE%\go\bin` está presente na variável de ambiente `PATH`.

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


## Comandos

O `mrtn` mantém uma interface minimalista e focada em produtividade:

### 1. Sincronizar casos de teste (`mrtn sync`)

Baixa ou atualiza o repositório contendo as questões, casos de entrada e gabaritos esperados ([`mrtn-oracle`](https://github.com/murilo-henrique060/mrtn-oracle)):

```bash
mrtn sync
```

> **Nota:** Os casos de teste sincronizados são obtidos de [murilo-henrique060/mrtn-oracle](https://github.com/murilo-henrique060/mrtn-oracle) e salvos no diretório de dados do usuário (`~/.local/share/mrtn/oracle` no Linux, `%LocalAppData%\mrtn\oracle` no Windows e `~/Library/Application Support/mrtn/oracle` no macOS, podendo ser alterado via `MRTN_DATA_DIR`), permitindo treinar e rodar testes mesmo sem conexão com a internet.

---

### 2. Consultar informações e enunciado (`mrtn info`)

Permite inspecionar as provas, questões, enunciados (`statement.md`) e casos de teste disponíveis no cache local:

```bash
mrtn info [prova] [questao] [numero-do-teste] [--web]
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

* **Informando prova e questão (`mrtn info <prova> <questao> [--web]`):**  
  Exibe os metadados da questão, o **enunciado completo** (a partir do arquivo `statement.md`, com fórmulas matemáticas LaTeX convertidas automaticamente para Unicode legível no terminal) e, logo abaixo, lista todos os casos de teste disponíveis numerados sequencialmente a partir de `1`.  
  Se a flag `--web` (ou `-w`) for informada, o `mrtn` também abre o enunciado renderizado com **KaTeX** no seu navegador padrão.
  ```bash
  mrtn info 2026-1 A
  mrtn info 2026-1 A --web
  ```

* **Informando prova, questão e número do teste (`mrtn info <prova> <questao> <numero-do-teste>`):**  
  Exibe informações detalhadas sobre o caso de teste específico, incluindo o conteúdo de entrada (`input`), saída esperada (`output`) e metadados.  
  > **Importante:** Apenas o número do teste (na ordem exibida pelo `mrtn info`, começando em `1`) é válido — por exemplo, use `1` ao invés de `sample001`.
  ```bash
  mrtn info 2026-1 A 1
  ```

### 3. Executar solução (`mrtn run`)

Executa o seu código contra os casos de teste da questão especificada, informando a prova e a questão (*case-insensitive*, ex: `2026-1 a` ou `2026-1 A`):

```bash
mrtn run <prova> <questao> [numero-do-teste] <arquivo>
```

> **Nota sobre `[numero-do-teste]`:** Ao informar o número de um teste específico (começando em `1`), o `mrtn run` exibe um diagnóstico focado para cada tipo de resultado:
> - **`AC` (Aceito):** Exibe apenas a linha de confirmação.
> - **`WA` (Resposta incorreta):** Exibe a entrada (caso não seja gigante), apenas a primeira linha de diferença entre o esperado e o obtido, eventuais prints de debug (`stderr`) e uma dica de depuração específica para a linguagem utilizada.
> - **`RTE` (Erro de execução):** Exibe a mensagem de erro (`stderr`) e os prints de debug (`stdout`) produzidos antes da falha.
> - **`TLE` (Tempo limite excedido):** Exibe os prints de debug capturados antes do estouro de tempo.

#### Exemplos de uso:

* **Rodar todos os casos da questão `A` da prova `2026-1`:**
  ```bash
  mrtn run 2026-1 A solucao.cpp
  ```

* **Rodar apenas um caso de teste específico para depuração (ex: caso `3` da questão `B` de `2026-1`):**
  ```bash
  mrtn run 2026-1 b 3 solucao.py
  ```

* **Rodar solução em Python para a questão `M` de `2026-1`:**
  ```bash
  mrtn run 2026-1 m solucao.py
  ```


### 4. Limpar dados locais (`mrtn clear`)

Remove os dados das provas, questões e casos de teste sincronizados no diretório de dados do usuário:

```bash
mrtn clear
```

> **Nota:** Após a limpeza, você pode baixar os casos de teste novamente a qualquer momento executando `mrtn sync`.


### 5. Atualizar a ferramenta (`mrtn update`)

Baixa e instala automaticamente a versão mais recente do `mrtn` a partir das releases do GitHub, substituindo o binário atual:

```bash
mrtn update
```

Você também pode especificar uma versão alvo ou forçar a reinstalação:

```bash
mrtn update v1.0.2
mrtn update --force
```


### 6. Desinstalar completamente (`mrtn remove`)

Remove todos os arquivos associados ao `mrtn` do sistema — incluindo casos de teste sincronizados, diretórios de dados locais, caches temporários de enunciados e o próprio executável do `mrtn`:

```bash
mrtn remove
```


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


## Licença

Distribuído sob a licença MIT. Consulte `LICENSE` para mais informações.
