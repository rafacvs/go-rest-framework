#!/bin/bash

# FRONTEND.sh - Interface interativa para API REST
# Framework Go - Gerenciamento de Usuários

# Cores para formatação (opcional)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações
API_BASE_URL="http://localhost:8080"
HAS_JQ=false

# Verificar dependências
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}Erro: curl não está instalado.${NC}"
        exit 1
    fi
    
    if command -v jq &> /dev/null; then
        HAS_JQ=true
    else
        echo -e "${YELLOW}Aviso: jq não está instalado. Usando parsing manual de JSON.${NC}"
        echo "Para melhor experiência, instale jq: sudo apt-get install jq (ou equivalente)"
    fi
}

# Função para fazer requisições HTTP
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    local url="${API_BASE_URL}${endpoint}"
    local response
    local status_code
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "$status_code|$response_body"
}

# Função para extrair campo de JSON
parse_json_field() {
    local json=$1
    local field=$2
    
    if [ "$HAS_JQ" = true ]; then
        echo "$json" | jq -r ".$field // empty" 2>/dev/null
    else
        # Parsing manual com grep/sed
        echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed "s/\"$field\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\"/\1/"
    fi
}

# Função para extrair campo numérico de JSON
parse_json_field_num() {
    local json=$1
    local field=$2
    
    if [ "$HAS_JQ" = true ]; then
        echo "$json" | jq -r ".$field // empty" 2>/dev/null
    else
        # Parsing manual para números
        echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*[0-9]*" | sed "s/\"$field\"[[:space:]]*:[[:space:]]*\([0-9]*\)/\1/"
    fi
}

# Função para extrair array de JSON (para list_users)
parse_json_array() {
    local json=$1
    
    if [ "$HAS_JQ" = true ]; then
        echo "$json" | jq -c '.[]' 2>/dev/null
    else
        # Parsing manual para array - extrai cada objeto
        # Remove colchetes externos e separa por objetos
        echo "$json" | sed 's/^\[//;s/\]$//' | sed 's/},{/}\n{/g' | sed 's/^{/{/;s/}$/}/'
    fi
}

# Função para formatar exibição de usuário
format_user_display() {
    local user_json=$1
    
    local id=$(parse_json_field_num "$user_json" "id")
    local nome=$(parse_json_field "$user_json" "nome")
    local email=$(parse_json_field "$user_json" "email")
    local telefone=$(parse_json_field "$user_json" "telefone")
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}ID:${NC}       $id"
    echo -e "${GREEN}Nome:${NC}     $nome"
    echo -e "${GREEN}Email:${NC}    $email"
    echo -e "${GREEN}Telefone:${NC} $telefone"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Função para listar todos os usuários
list_users() {
    echo -e "\n${BLUE}Listando usuários...${NC}\n"
    
    local result=$(make_request "GET" "/users")
    local status_code=$(echo "$result" | cut -d'|' -f1)
    local response_body=$(echo "$result" | cut -d'|' -f2-)
    
    if [ "$status_code" -eq 200 ]; then
        if [ "$HAS_JQ" = true ]; then
            local count=$(echo "$response_body" | jq 'length' 2>/dev/null)
            if [ "$count" -eq 0 ]; then
                echo -e "${YELLOW}Nenhum usuário encontrado.${NC}\n"
                return
            fi
            
            echo "$response_body" | jq -r '.[] | "\(.id)|\(.nome)|\(.email)|\(.telefone)"' | while IFS='|' read -r id nome email telefone; do
                echo -e "${GREEN}[$id]${NC} $nome - $email - $telefone"
            done
        else
            # Parsing manual
            local users=$(parse_json_array "$response_body")
            if [ -z "$users" ]; then
                echo -e "${YELLOW}Nenhum usuário encontrado.${NC}\n"
                return
            fi
            
            echo "$users" | while read -r user; do
                local id=$(parse_json_field_num "$user" "id")
                local nome=$(parse_json_field "$user" "nome")
                local email=$(parse_json_field "$user" "email")
                local telefone=$(parse_json_field "$user" "telefone")
                echo -e "${GREEN}[$id]${NC} $nome - $email - $telefone"
            done
        fi
    else
        local error_msg=$(parse_json_field "$response_body" "error")
        echo -e "${RED}Erro: ${error_msg:-"Erro desconhecido"}${NC}"
    fi
    
    echo ""
}

# Função para buscar usuário por ID
get_user_by_id() {
    echo -e "\n${BLUE}Buscar usuário por ID${NC}"
    read -p "Digite o ID do usuário: " user_id
    
    # Validação
    if ! [[ "$user_id" =~ ^[0-9]+$ ]]; then
        echo -e "${RED}Erro: ID deve ser um número.${NC}\n"
        return
    fi
    
    echo ""
    local result=$(make_request "GET" "/users/$user_id")
    local status_code=$(echo "$result" | cut -d'|' -f1)
    local response_body=$(echo "$result" | cut -d'|' -f2-)
    
    if [ "$status_code" -eq 200 ]; then
        format_user_display "$response_body"
    else
        local error_msg=$(parse_json_field "$response_body" "error")
        echo -e "${RED}Erro: ${error_msg:-"Erro desconhecido"}${NC}"
    fi
    
    echo ""
}

# Função para editar usuário
edit_user() {
    echo -e "\n${BLUE}Editar usuário${NC}"
    read -p "Digite o ID do usuário: " user_id
    
    # Validação
    if ! [[ "$user_id" =~ ^[0-9]+$ ]]; then
        echo -e "${RED}Erro: ID deve ser um número.${NC}\n"
        return
    fi
    
    echo -e "\n${YELLOW}Buscando usuário...${NC}"
    local result=$(make_request "GET" "/users/$user_id")
    local status_code=$(echo "$result" | cut -d'|' -f1)
    local response_body=$(echo "$result" | cut -d'|' -f2-)
    
    if [ "$status_code" -ne 200 ]; then
        local error_msg=$(parse_json_field "$response_body" "error")
        echo -e "${RED}Erro: ${error_msg:-"Erro desconhecido"}${NC}\n"
        return
    fi
    
    # Extrair valores atuais
    local current_nome=$(parse_json_field "$response_body" "nome")
    local current_email=$(parse_json_field "$response_body" "email")
    local current_telefone=$(parse_json_field "$response_body" "telefone")
    
    echo -e "\n${BLUE}Usuário atual:${NC}"
    format_user_display "$response_body"
    
    echo -e "\n${YELLOW}Digite os novos valores (pressione Enter para manter o valor atual):${NC}\n"
    
    # Editar nome
    read -p "Nome [$current_nome]: " new_nome
    if [ -z "$new_nome" ]; then
        new_nome="$current_nome"
    fi
    
    # Editar email
    read -p "Email [$current_email]: " new_email
    if [ -z "$new_email" ]; then
        new_email="$current_email"
    fi
    
    # Editar telefone
    read -p "Telefone [$current_telefone]: " new_telefone
    if [ -z "$new_telefone" ]; then
        new_telefone="$current_telefone"
    fi
    
    # Verificar se houve alterações
    if [ "$new_nome" = "$current_nome" ] && [ "$new_email" = "$current_email" ] && [ "$new_telefone" = "$current_telefone" ]; then
        echo -e "${YELLOW}Nenhuma alteração detectada. Operação cancelada.${NC}\n"
        return
    fi
    
    # Criar JSON para atualização
    local json_data
    if [ "$HAS_JQ" = true ]; then
        json_data=$(jq -n \
            --arg nome "$new_nome" \
            --arg email "$new_email" \
            --arg telefone "$new_telefone" \
            '{nome: $nome, email: $email, telefone: $telefone}')
    else
        # Criar JSON manualmente
        json_data="{\"nome\":\"$new_nome\",\"email\":\"$new_email\",\"telefone\":\"$new_telefone\"}"
    fi
    
    echo -e "\n${YELLOW}Atualizando usuário...${NC}"
    result=$(make_request "PUT" "/users/$user_id" "$json_data")
    status_code=$(echo "$result" | cut -d'|' -f1)
    response_body=$(echo "$result" | cut -d'|' -f2-)
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}Usuário atualizado com sucesso!${NC}\n"
        format_user_display "$response_body"
    else
        local error_msg=$(parse_json_field "$response_body" "error")
        echo -e "${RED}Erro: ${error_msg:-"Erro desconhecido"}${NC}"
    fi
    
    echo ""
}

# Função para excluir usuário
delete_user() {
    echo -e "\n${BLUE}Excluir usuário${NC}"
    read -p "Digite o ID do usuário: " user_id
    
    # Validação
    if ! [[ "$user_id" =~ ^[0-9]+$ ]]; then
        echo -e "${RED}Erro: ID deve ser um número.${NC}\n"
        return
    fi
    
    # Buscar usuário para mostrar antes de deletar
    echo -e "\n${YELLOW}Buscando usuário...${NC}"
    local result=$(make_request "GET" "/users/$user_id")
    local status_code=$(echo "$result" | cut -d'|' -f1)
    local response_body=$(echo "$result" | cut -d'|' -f2-)
    
    if [ "$status_code" -ne 200 ]; then
        local error_msg=$(parse_json_field "$response_body" "error")
        echo -e "${RED}Erro: ${error_msg:-"Erro desconhecido"}${NC}\n"
        return
    fi
    
    format_user_display "$response_body"
    
    echo -e "${RED}ATENÇÃO: Esta ação não pode ser desfeita!${NC}"
    read -p "Tem certeza que deseja excluir este usuário? (s/N): " confirm
    
    if [ "$confirm" != "s" ] && [ "$confirm" != "S" ]; then
        echo -e "${YELLOW}Operação cancelada.${NC}\n"
        return
    fi
    
    echo -e "\n${YELLOW}Excluindo usuário...${NC}"
    result=$(make_request "DELETE" "/users/$user_id")
    status_code=$(echo "$result" | cut -d'|' -f1)
    response_body=$(echo "$result" | cut -d'|' -f2-)
    
    if [ "$status_code" -eq 200 ]; then
        local message=$(parse_json_field "$response_body" "message")
        echo -e "${GREEN}${message:-"Usuário excluído com sucesso!"}${NC}\n"
    else
        local error_msg=$(parse_json_field "$response_body" "error")
        echo -e "${RED}Erro: ${error_msg:-"Erro desconhecido"}${NC}\n"
    fi
}

# Menu principal
show_menu() {
    echo -e "\n${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   Gerenciamento de Usuários - API      ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo "1) Listar usuários"
    echo "2) Buscar usuário por ID"
    echo "3) Editar usuário"
    echo "4) Excluir usuário"
    echo "5) Sair"
    echo ""
}

# Função principal
main() {
    check_dependencies
    
    echo -e "${GREEN}Interface interativa iniciada!${NC}"
    echo -e "${YELLOW}API: $API_BASE_URL${NC}\n"
    
    while true; do
        show_menu
        read -p "Escolha uma opção: " choice
        
        case $choice in
            1)
                list_users
                ;;
            2)
                get_user_by_id
                ;;
            3)
                edit_user
                ;;
            4)
                delete_user
                ;;
            5)
                echo -e "${GREEN}Saindo...${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Opção inválida. Tente novamente.${NC}\n"
                ;;
        esac
        
        read -p "Pressione Enter para continuar..."
    done
}

# Executar função principal
main

