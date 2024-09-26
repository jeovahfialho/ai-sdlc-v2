import sys
from openai import OpenAI

client = OpenAI(api_key="")
import os
import yaml

# Global variable to store conversation context
conversation_context = []

def send_to_chatgpt(prompt, file_extension, prompt_file_content, yaml_content):
    try:
        file_name = prompt.split(' ')[-1]
        print(f"Sending prompt to ChatGPT for file: {file_name}")

        # Determine the language based on the file extension
        language = ""
        if file_extension == ".go":
            language = "go"
        elif file_extension == ".yml" or file_extension == ".yaml":
            language = "yaml"
            prompt += yaml_content
        elif file_extension == ".sql":
            language = "sql"
        elif file_extension == "Dockerfile":
            language = "dockerfile"
        elif file_name.lower() == "readme.md":
            language = "markdown"
            prompt += "\nPlease include detailed instructions on how to test the APIs using curl commands."
        else:
            language = "plaintext"

        # Combine prompt file content with the current prompt
        combined_prompt = f"{prompt_file_content}\n{prompt}"

        # Format the prompt to ask for content within code blocks
        formatted_prompt = f"Provide the content for {file_name} in {language}:\n```{language}\n{combined_prompt}\n```"

        # Update the conversation context
        conversation_context.append({"role": "user", "content": formatted_prompt})

        response = client.chat.completions.create(model="gpt-3.5-turbo",
        messages=conversation_context)
        print(f"Received response from ChatGPT for file: {file_name}")

        # Extract content within the first ``` block only
        content = response.choices[0].message.content
        start_index = content.find("```")
        if start_index != -1:
            start_index += 3  # Move past the ```
            end_index = content.find("```", start_index)
            if end_index != -1:
                content = content[start_index:end_index]

        # Remove the first line if it matches the file extension
        lines = content.strip().split('\n')
        if lines[0].strip() == language:
            lines = lines[1:]

        # Update the context with the assistant's response
        conversation_context.append({"role": "assistant", "content": content.strip()})

        return "\n".join(lines).strip()

    except Exception as e:
        print(f"Error sending prompt to ChatGPT: {e}")
        sys.exit(1)

def create_files_from_yaml(yaml_content, prompt_file_content):
    try:
        print("Creating files and directories from YAML content...")
        yaml_structure = yaml.safe_load(yaml_content)

        def create_structure(base_path, structure):
            for key, value in structure.items():
                if isinstance(value, list):  # Arquivos dentro de uma pasta
                    dir_path = os.path.join(base_path, key)
                    os.makedirs(dir_path, exist_ok=True)
                    print(f"Created directory: {dir_path}")
                    for file in value:
                        file_path = os.path.join(dir_path, file)
                        with open(file_path, 'w', encoding='utf-8') as f:
                            pass  # Cria arquivo vazio
                        print(f"Created file: {file_path}")
                elif isinstance(value, dict):  # Subdiretórios
                    dir_path = os.path.join(base_path, key)
                    os.makedirs(dir_path, exist_ok=True)
                    print(f"Created directory: {dir_path}")
                    create_structure(dir_path, value)

        # Adjust base path to move one level up (to the same level as backend and frontend)
        base_path = os.path.abspath(os.path.join('.', '..'))
        create_structure(base_path, yaml_structure)
    except yaml.YAMLError as e:
        print(f"YAML parsing error: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"Error creating files and directories: {e}")
        sys.exit(1)

def update_files_with_content(yaml_content, prompt_file_content):
    try:
        print("Updating files with content from ChatGPT...")
        yaml_structure = yaml.safe_load(yaml_content)

        def update_structure(base_path, structure):
            for key, value in structure.items():
                if isinstance(value, list):  # Arquivos dentro de uma pasta
                    dir_path = os.path.join(base_path, key)
                    for file in value:
                        file_path = os.path.join(dir_path, file)
                        # Extract the file extension or set as Dockerfile
                        file_extension = os.path.splitext(file)[1] if "." in file else "Dockerfile"
                        prompt = f"Provide the content for {file} in {dir_path}."
                        content = send_to_chatgpt(prompt, file_extension, prompt_file_content, yaml_content)
                        with open(file_path, 'w', encoding='utf-8') as f:
                            f.write(content)
                        print(f"Updated file: {file_path} with content from ChatGPT.")
                elif isinstance(value, dict):  # Subdiretórios
                    dir_path = os.path.join(base_path, key)
                    update_structure(dir_path, value)

        # Adjust base path to move one level up (to the same level as backend and frontend)
        base_path = os.path.abspath(os.path.join('.', '..'))
        update_structure(base_path, yaml_structure)
    except yaml.YAMLError as e:
        print(f"YAML parsing error: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"Error updating files with content: {e}")
        sys.exit(1)

def main():
    if len(sys.argv) < 3:
        print("Usage: python create_project_files.py <yaml_file> <prompt_file>")
        sys.exit(1)

    yaml_file = sys.argv[1]
    prompt_file = sys.argv[2]

    print(f"Reading YAML content from file: {yaml_file}")
    try:
        with open(yaml_file, 'r', encoding='utf-8') as file:
            yaml_content = file.read()
            print("Successfully read the YAML content.")
    except Exception as e:
        print(f"Error reading YAML file: {e}")
        sys.exit(1)

    print(f"Reading prompt content from file: {prompt_file}")
    try:
        with open(prompt_file, 'r', encoding='utf-8') as file:
            prompt_file_content = file.read()
            print("Successfully read the prompt content.")
    except Exception as e:
        print(f"Error reading prompt file: {e}")
        sys.exit(1)

    # 1. Criar diretórios e arquivos vazios conforme YAML
    create_files_from_yaml(yaml_content, prompt_file_content)

    # 2. Atualizar os arquivos com o conteúdo do ChatGPT
    update_files_with_content(yaml_content, prompt_file_content)

    print("Script completed successfully.")

if __name__ == "__main__":
    main()
