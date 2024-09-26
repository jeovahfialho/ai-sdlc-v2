import openai
import os

# Defina sua chave de API da OpenAI
openai.api_key = os.getenv("OPENAI_API_KEY")

def generate_code_from_prompt(prompt, file_name="generated_system.py"):
    """
    Gera código de um sistema completo a partir de um prompt utilizando a API da OpenAI.
    :param prompt: Descrição do sistema em formato de string.
    :param file_name: Nome do arquivo onde o código será salvo.
    """
    try:
        # Envia a requisição para o modelo GPT-4 da OpenAI
        response = openai.ChatCompletion.create(
            model="gpt-4",  # Usando o modelo GPT-4
            messages=[{"role": "user", "content": prompt}],  # Enviando o prompt como uma mensagem
            max_tokens=3000,  # Ajuste conforme necessário para garantir uma resposta longa o suficiente
            temperature=0.5  # Controla a criatividade do código
        )

        # Extrai o código gerado da resposta
        code = response['choices'][0]['message']['content'].strip()

        # Salva o código em um arquivo
        with open(file_name, 'w') as code_file:
            code_file.write(code)

        print(f"Código gerado e salvo no arquivo {file_name}")

    except Exception as e:
        print(f"Ocorreu um erro: {e}")

if __name__ == "__main__":
    # Exemplo de uso - prompt para gerar um sistema completo de notificações para um app de cassino
    prompt_description = """
    Crie o código completo de um sistema para um aplicativo de cassino que envie notificações em tempo real aos usuários. O sistema deve incluir:
    
    1. **Tipos de Notificações**:
       - Atualizações de pontos: Informe os usuários sobre novos pontos de fidelidade ganhos.
       - Promoções especiais: Alerta os usuários sobre ofertas e bônus exclusivos.
       - Disponibilidade de jogos: Notifique os usuários quando seus jogos favoritos estiverem disponíveis.

    2. **Interface do Usuário**:
       - Banner não intrusivo no topo da tela.
       - Expansível para mais detalhes.
       - Aparência personalizável (esquema de cores, tamanho do texto).

    3. **Funcionalidade**:
       - Enviar notificações push para o dispositivo do usuário.
       - Notificações dentro do app.
       - Clicável para ver mais detalhes ou tomar ações imediatas.

    4. **Personalização**:
       - Os usuários podem escolher quais tipos de notificações receber.
       - Controles de frequência para gerenciar o volume de notificações.
       - Configuração de "horas silenciosas" para pausar notificações durante horários especificados.

    5. **Recursos de Acessibilidade**:
       - Modo de alto contraste.
       - Tamanho do texto ajustável.
       - Opção de feedback tátil.

    6. **Integração**:
       - Conexão perfeita com recursos existentes do app (saldo de pontos, lobby de jogos).
       - Links diretos para seções relevantes do app a partir das notificações.

    7. **Controle do Usuário**:
       - Fácil opção de opt-in/opt-out.
       - Capacidade de adiar ou descartar notificações.

    8. **Análises**:
       - Acompanhe o engajamento dos usuários com as notificações.
       - Meça o impacto no uso do app e nas visitas ao cassino.
    """

    # Gera o código com base no prompt e salva no arquivo especificado
    generate_code_from_prompt(prompt_description, "casino_notification_system.py")
