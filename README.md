# Documentação do Framework GoFramework

## Visão Geral

GoFramework é um framework REST simples em Go que permite criar APIs através de uma estrutura de diretórios. O framework converte automaticamente a estrutura de pastas em rotas HTTP, onde cada arquivo representa um método HTTP e cada pasta representa um segmento da URL.

## Estrutura de Arquivos

```
pkg/goframework/
├── app.go          # Aplicação principal e gerenciamento de rotas
├── router.go       # Sistema de roteamento e matching
├── context.go      # Contexto da requisição HTTP
├── registry.go     # Registro de handlers durante init()
└── types.go        # Tipos e erros HTTP
```

## Pré-requisitos

Para o desenvolvimento do GoFramework, utilizamos a linguagem Go, acompanhada de suas bibliotecas padrão para criação de servidores HTTP, manipulação de JSON e controle de concorrência. Além disso, utilizamos um ambiente de desenvolvimento compatível com Go 1.21+ e um sistema de arquivos local para organização da estrutura de rotas.

## Instalação

Para executar o projeto, é necessário ter instalado o **Go 1.25.3**. Abaixo segue um passo a passo rápido para realizar a instalação:

1. **Acesse o site oficial do Go**  
   Entre em: https://go.dev/dl e baixe o instalador correspondente ao seu sistema operacional (Windows, Linux ou macOS).

2. **Execute o instalador**  
   No Windows ou macOS, basta seguir o assistente padrão de instalação.  
   No Linux, extraia o arquivo `.tar.gz` para `/usr/local` utilizando:  
   ```bash
   sudo tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz
   ```

3. **Configure a variável de ambiente `PATH`**  
   Adicione o diretório do Go ao PATH:  
   ```bash
   export PATH=$PATH:/usr/local/go/bin
   ```

4. **Verifique a instalação**  
   Execute o comando:  
   ```bash
   go version
   ```  
   Ele deve exibir `go version go1.25.3 ...`.

Com isso, o ambiente estará pronto para rodar o projeto.


## Componentes Principais

### 1. types.go - Tipos Base

Define os tipos fundamentais do framework.

**HandlerFunc**
- Tipo: `func(*Context) error`
- Assinatura padrão para todas as funções de handler
- Recebe um Context e retorna um error (nil se sucesso)

**HTTPError**
- Estrutura para erros HTTP customizados
- Campos: `Status int`, `Message string`
- Implementa a interface `error`

**NewHTTPError(status int, message string) error**
- Cria um erro HTTP com status e mensagem específicos
- Usado pelos handlers para retornar erros HTTP

### 2. registry.go - Sistema de Registro

Gerencia o registro de handlers durante a inicialização do programa.

**Variáveis globais:**
- `routesReg`: mapa que armazena todas as rotas registradas
- `registryMu`: mutex para sincronização thread-safe

**Register(handler HandlerFunc)**
- Função chamada pelos handlers em suas funções `init()`
- Usa `runtime.Caller(1)` para obter o caminho do arquivo que chamou
- Extrai o método HTTP do nome do arquivo (get.go = GET, post.go = POST, etc)
- Armazena o handler no mapa global `routesReg` usando o caminho absoluto do arquivo como chave
- Thread-safe através de mutex

**listRegistrations() []registration**
- Retorna lista de todas as rotas registradas
- Usado por `LoadRoutes()` para carregar as rotas no router
- Thread-safe

**methodFromFilename(path string) (string, error)**
- Extrai o método HTTP do nome do arquivo
- Suporta: get, post, put, patch, delete, options
- Retorna erro se o nome não for reconhecido

### 3. router.go - Sistema de Roteamento

Gerencia o roteamento de requisições HTTP para os handlers corretos.

**Estruturas:**

`route`
- Representa uma rota registrada
- Campos: `method string`, `pattern string`, `segments []routeSegment`, `handler HandlerFunc`

`routeSegment`
- Representa um segmento de uma rota (parte da URL)
- Campos: `value string`, `isParam bool`, `paramKey string`
- `isParam` indica se é um parâmetro dinâmico (ex: `:id`)

`Router`
- Gerencia todas as rotas
- Campo: `routes []route`

**NewRouter() *Router**
- Cria uma nova instância do Router
- Inicializa slice vazio de rotas

**AddRoute(method, pattern string, handler HandlerFunc) error**
- Adiciona uma rota ao router
- Valida método, pattern e handler
- Normaliza o método para maiúsculas
- Normaliza o pattern através de `normalizePattern()`
- Verifica duplicatas (mesmo método + pattern)
- Converte o pattern em segmentos através de `buildSegments()`
- Adiciona a rota ao slice

**Match(method, path string) (HandlerFunc, map[string]string, bool)**
- Busca uma rota que corresponda ao método e caminho
- Normaliza método para maiúsculas
- Converte o path em segmentos através de `splitPath()`
- Itera sobre todas as rotas procurando correspondência
- Para cada rota, chama `matchSegments()` para verificar correspondência
- Retorna handler, parâmetros extraídos e true se encontrou
- Retorna nil, nil, false se não encontrou

**matchSegments(patternSegs, pathSegs []routeSegment) (map[string]string, bool)**
- Compara segmentos do pattern com segmentos do path
- Verifica se têm o mesmo tamanho
- Para cada segmento do pattern:
  - Se `isParam` é true, extrai o valor do path e adiciona ao mapa de parâmetros
  - Se não é parâmetro, compara valores exatamente
- Retorna mapa de parâmetros e true se todos correspondem
- Retorna nil, false se não corresponde

**buildSegments(pattern string) ([]routeSegment, error)**
- Converte um pattern string em slice de routeSegment
- Primeiro chama `splitPath()` para obter segmentos básicos
- Para cada segmento:
  - Se começa com `:`, cria um routeSegment com `isParam = true`
  - Caso contrário, cria segmento normal
- Valida que parâmetros têm nome válido (não vazio após `:`)

**splitPath(path string) []routeSegment**
- Divide um caminho em segmentos
- Normaliza o path primeiro
- Se path é `/`, retorna slice vazio
- Remove `/` inicial e divide por `/`
- Cria routeSegment para cada parte (ainda sem identificar parâmetros)

**normalizePattern(pattern string) string**
- Normaliza um pattern de rota
- Se vazio, retorna `/`
- Remove espaços
- Garante que começa com `/`
- Remove `/` final (exceto se for apenas `/`)
- Retorna pattern normalizado

### 4. context.go - Contexto da Requisição

Fornece acesso aos dados da requisição HTTP e métodos para resposta.

**Estruturas:**

`Context`
- Contexto passado para cada handler
- Campos:
  - `Req *http.Request`: requisição HTTP original
  - `Res http.ResponseWriter`: response writer original
  - `Params map[string]string`: parâmetros extraídos da URL
  - `writer *responseWriter`: wrapper do ResponseWriter
  - `bodyOnce sync.Once`: garante leitura única do body
  - `body []byte`: corpo da requisição em cache
  - `bodyErr error`: erro na leitura do body

`responseWriter`
- Wrapper do `http.ResponseWriter` padrão
- Campos: `ResponseWriter http.ResponseWriter`, `wrote bool`
- `wrote` indica se headers já foram escritos

**newContext(w *responseWriter, r *http.Request, params map[string]string) *Context**
- Construtor privado do Context
- Cria cópia do mapa de parâmetros
- Inicializa Context com requisição, response writer e parâmetros

**Context.JSON(status int, data any) error**
- Escreve resposta JSON
- Verifica se já escreveu headers (retorna nil se sim)
- Define Content-Type como application/json
- Escreve status code
- Codifica data como JSON e escreve no response writer

**Context.Param(name string) string**
- Retorna valor de parâmetro da URL por nome
- Exemplo: para rota `/users/:id`, `Param("id")` retorna o valor

**Context.Query() url.Values**
- Retorna query parameters da URL
- Wrapper direto para `Req.URL.Query()`

**Context.QueryParam(name string) string**
- Retorna valor de um query parameter específico
- Wrapper para `Query().Get(name)`

**Context.SetHeader(name, value string)**
- Define um header HTTP na resposta
- Só funciona se headers ainda não foram escritos
- Wrapper para `Header().Set()`

**Context.Body() ([]byte, error)**
- Lê o corpo da requisição HTTP
- Usa `sync.Once` para garantir leitura única
- Fecha o body após leitura
- Cacheia o resultado em `body` e `bodyErr`
- Retorna o mesmo resultado em chamadas subsequentes

**responseWriter.WriteHeader(statusCode int)**
- Escreve status code HTTP
- Previne escrita múltipla através da flag `wrote`
- Chama `ResponseWriter.WriteHeader()` do Go padrão

**responseWriter.Write(b []byte) (int, error)**
- Escreve dados na resposta
- Se headers não foram escritos, escreve StatusOK automaticamente
- Delega para `ResponseWriter.Write()` do Go padrão

### 5. app.go - Aplicação Principal

Gerencia a aplicação, carregamento de rotas e tratamento de requisições HTTP.

**Estrutura:**

`App`
- Aplicação principal do framework
- Campos:
  - `router *Router`: instância do router
  - `routesDir string`: diretório raiz das rotas

**New() *App**
- Construtor da aplicação
- Cria novo Router através de `NewRouter()`
- Retorna instância de App inicializada

**App.LoadRoutes(root string) error**
- Carrega rotas do sistema de arquivos
- Valida que root não está vazio
- Converte root para caminho absoluto
- Chama `listRegistrations()` para obter rotas registradas
- Valida que existem rotas registradas
- Para cada rota registrada:
  - Verifica se o arquivo está dentro do diretório root
  - Converte caminho do arquivo em pattern HTTP através de `buildPatternFromFile()`
  - Adiciona rota ao router através de `router.AddRoute()`
- Valida que pelo menos uma rota foi carregada
- Armazena routesDir para referência futura

**App.Listen(addr string) error**
- Inicia servidor HTTP
- Valida que router está inicializado
- Chama `http.ListenAndServe()` passando addr e a própria App (que implementa http.Handler)
- Bloqueia até erro ou interrupção

**App.ServeHTTP(w http.ResponseWriter, r *http.Request)**
- Implementa interface `http.Handler`
- Chamado automaticamente pelo servidor HTTP do Go para cada requisição
- Fluxo:
  1. Chama `router.Match()` para encontrar handler correspondente
  2. Se não encontrou, retorna erro 404 JSON
  3. Cria `responseWriter` wrapper
  4. Cria `Context` através de `newContext()`
  5. Chama o handler passando o Context
  6. Se handler retornou erro e resposta ainda não foi escrita, chama `handleError()`

**App.handleError(w http.ResponseWriter, err error)**
- Trata erros retornados pelos handlers
- Se erro é `*HTTPError`, retorna status e mensagem do erro
- Caso contrário, retorna erro 500 genérico
- Usa `writeJSONError()` para escrever resposta JSON

**writeJSONError(w http.ResponseWriter, status int, message string)**
- Função auxiliar para escrever erros em formato JSON
- Cria objeto `{"error": message}`
- Define Content-Type como application/json
- Escreve status code
- Codifica e escreve JSON

**buildPatternFromFile(root, file string) (string, error)**
- Converte caminho de arquivo em pattern HTTP
- Calcula caminho relativo do arquivo em relação ao root
- Valida que arquivo está dentro do root (não permite `..`)
- Obtém diretório do arquivo relativo
- Se diretório é `.`, trata como vazio
- Divide diretório em segmentos
- Para cada segmento:
  - Ignora segmentos vazios ou `index`
  - Se segmento começa com `_`, trata como parâmetro (converte `_id` para `:id`)
  - Caso contrário, adiciona como segmento normal
- Junta segmentos com `/` e adiciona `/` inicial
- Retorna pattern (ex: `/users/:id`)

### 6. FRONTEND.sh - Interface

O frontend do projeto consiste em um script Bash interativo responsável por consumir a API e oferecer ao usuário uma interface simples via terminal. Ele utiliza curl para enviar requisições HTTP e, quando disponível, o utilitário jq para interpretar respostas em JSON. Por meio deste script, é possível listar usuários, buscar por ID, editar e excluir registros, tudo de forma rápida e intuitiva, sem necessidade de abrir um navegador ou ferramentas externas.

## Fluxo de Execução

### 1. Inicialização do Programa

```
main() executa
  ↓
Imports executam funções init() dos pacotes de rotas
  ↓
Cada init() chama goframework.Register(handler)
  ↓
Register() armazena handler no mapa global routesReg
  ↓
main() chama app.New()
  ↓
New() cria Router vazio
  ↓
main() chama app.LoadRoutes("example/app/routes")
  ↓
LoadRoutes() chama listRegistrations()
  ↓
Para cada rota registrada:
  - buildPatternFromFile() converte caminho em pattern
  - router.AddRoute() adiciona rota ao router
  ↓
main() chama app.Listen(":8080")
  ↓
Listen() inicia servidor HTTP
```

### 2. Processamento de Requisição HTTP

```
Cliente envia requisição HTTP
  ↓
Go chama App.ServeHTTP(w, r)
  ↓
ServeHTTP() chama router.Match(method, path)
  ↓
Match() itera sobre rotas:
  - Para cada rota, chama matchSegments()
  - Se corresponde, retorna handler e parâmetros
  ↓
Se não encontrou, ServeHTTP() retorna 404 JSON
  ↓
Se encontrou:
  - Cria responseWriter wrapper
  - Cria Context com requisição e parâmetros
  - Chama handler(ctx)
  ↓
Handler executa lógica:
  - Pode ler ctx.Body() para obter dados
  - Pode ler ctx.Param() para obter parâmetros
  - Pode ler ctx.QueryParam() para query params
  - Chama ctx.JSON() para retornar resposta
  ↓
Se handler retornou erro:
  - ServeHTTP() chama handleError()
  - handleError() escreve erro JSON
```

### 3. Exemplo: Requisição GET /users/1

```
Requisição: GET /users/1
  ↓
ServeHTTP() recebe w, r
  ↓
router.Match("GET", "/users/1")
  ↓
Match() itera rotas:
  - Rota 1: GET /users
    - matchSegments() compara ["users"] com ["users", "1"]
    - Não corresponde (tamanhos diferentes)
  - Rota 2: GET /users/:id
    - matchSegments() compara [users, :id] com [users, 1]
    - Segmento 1: "users" == "users" ✓
    - Segmento 2: :id é parâmetro, extrai "1"
    - Retorna handler, params={"id": "1"}, true
  ↓
ServeHTTP() cria Context com params
  ↓
Handler Get(ctx) executa:
  - ctx.Param("id") retorna "1"
  - Converte "1" para int
  - Busca usuário no serviço
  - ctx.JSON(200, user) escreve resposta
  ↓
Resposta enviada ao cliente
```
## Exemplo de execução 

Em primeiro lugar a API desenvolvida com o GoFramework é iniciada. O servidor é carregado, as rotas são registradas e a aplicação fica pronta para receber requisições HTTP.

![Imagem](https://github.com/rafacvs/go-rest-framework/blob/master/imagem1.png)

Em seguida, é inicializado o frontend.

![Imagem](https://github.com/rafacvs/go-rest-framework/blob/master/imagem2.png)

E por fim, a execução da aplicação já em funcionamento, exibindo a listagem de usuários obtida a partir da API.

![Imagem](https://github.com/rafacvs/go-rest-framework/blob/master/imagem3.png)

## Convenções de Estrutura de Arquivos

### Mapeamento de Arquivos para Rotas

- `routes/index/get.go` → `GET /`
- `routes/users/get.go` → `GET /users`
- `routes/users/post.go` → `POST /users`
- `routes/users/_id/get.go` → `GET /users/:id`
- `routes/users/_id/put.go` → `PUT /users/:id`
- `routes/users/_id/delete.go` → `DELETE /users/:id`

### Regras

1. Nome do arquivo determina método HTTP: `get.go` = GET, `post.go` = POST, etc.
2. Estrutura de pastas determina o caminho da URL
3. Pasta `index` é ignorada (vira `/`)
4. Pastas começando com `_` viram parâmetros (ex: `_id` vira `:id`)
5. Cada arquivo deve ter função `init()` que chama `goframework.Register(HandlerFunc)`
6. Handler deve ter assinatura `func(*Context) error`

## Exemplo de Handler

```go
package users

import (
    "my-go-framework/pkg/goframework"
)

func init() {
    goframework.Register(Get)
}

func Get(ctx *goframework.Context) error {
    // Ler parâmetros da URL
    id := ctx.Param("id")
    
    // Ler query parameters
    page := ctx.QueryParam("page")
    
    // Ler corpo da requisição (se necessário)
    body, err := ctx.Body()
    
    // Retornar resposta JSON
    return ctx.JSON(200, map[string]string{"id": id})
}
```

## Tratamento de Erros

Handlers podem retornar erros de duas formas:

1. **HTTPError customizado:**
   ```go
   return goframework.NewHTTPError(404, "usuário não encontrado")
   ```
   - Retorna status HTTP específico com mensagem

2. **Erro genérico:**
   ```go
   return fmt.Errorf("erro interno")
   ```
   - Retorna status 500 com mensagem "internal_error"

O framework automaticamente:
- Detecta se resposta já foi escrita (não sobrescreve)
- Converte erros em respostas JSON
- Trata erros HTTPError com status correto

## Thread Safety

- `registry.go` usa mutex para proteger o mapa de rotas durante registro
- `context.go` usa `sync.Once` para garantir leitura única do body
- `responseWriter` previne escrita múltipla de headers
- Router é thread-safe para leitura (rotas não mudam após LoadRoutes)

## Limitações e Considerações

1. Rotas são carregadas uma vez na inicialização (não suporta hot-reload)
2. Matching de rotas é linear (O(n) onde n é número de rotas)
3. Parâmetros de rota são sempre strings (conversão manual necessária)
4. Body da requisição é lido completamente na memória
5. Não há middleware system nativo
6. Não há validação automática de dados
7. Estrutura de diretórios deve seguir convenções estritas

## Autores

Lucas Alves Zito
Rafael de Campos Villa da Silveira

