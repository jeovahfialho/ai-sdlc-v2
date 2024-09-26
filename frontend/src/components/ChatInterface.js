import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Container, TextField, Button, List, ListItem, ListItemText, Paper, Typography, Box } from '@mui/material';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { coy as codeStyle } from 'react-syntax-highlighter/dist/esm/styles/prism';

const ChatInterface = () => {
    const [message, setMessage] = useState("");
    const [chat, setChat] = useState([]);
    const [step, setStep] = useState(1); // Step inicial
    const [isChatBlocked, setIsChatBlocked] = useState(false); // Novo estado para bloquear o chat

    useEffect(() => {
        console.log("Initial welcome message sent.");
        setChat([{ user: "ChatApp v2", response: "Hello, I am the ChatApp v2. I am here to help you develop and deploy your project. Tell me, what exactly would you like to build?" }]);
    }, []);

    // Função que envia a mensagem para o backend
    const sendMessage = async (msg) => {
        try {
            console.log("Sending message to server:", msg);
            const response = await axios.post("http://localhost:8000/api/chat", { message: msg, step: step });
            const formattedResponse = formatResponse(response.data.response);

            console.log("Formatted response array:", formattedResponse);

            setChat(prevChat => [...prevChat, { user: "You", response: msg }, { user: "ChatApp v2", response: formattedResponse }]);
            setMessage("");

            // Atualiza o step baseado na resposta do backend
            const newStep = response.data.step;
            setStep(newStep);
            console.log("Step updated to:", newStep);

            // Verifica se a resposta contém um link e bloqueia o chat
            if (response.data.response.includes("<a href='")) {
                setIsChatBlocked(true); // Bloqueia o chat para impedir novas mensagens
            }

        } catch (error) {
            console.error("Error sending message:", error);
            setChat(prevChat => [...prevChat, { user: "You", response: msg }, { user: "ChatApp v2", response: "Error connecting to the server." }]);
            setMessage("");
        }
    };

    // Função que detecta links e formata YAML
    const formatResponse = (response) => {
        console.log("Formatting response:", response);
        
        const yamlMatch = response.match(/```yaml([\s\S]*?)```/);
        let formattedResponse = [];

        // Verifica se a resposta contém YAML
        if (yamlMatch) {
            console.log("YAML content detected.");
            formattedResponse.push(
                <SyntaxHighlighter language="yaml" style={codeStyle} key="yaml">
                    {yamlMatch[1]}
                </SyntaxHighlighter>
            );
        }

        // Remove o YAML da resposta para pegar o texto restante
        const textResponse = response.replace(/```yaml[\s\S]*?```/, '').trim();
        if (textResponse) {
            const lines = textResponse.split('\n');
            formattedResponse = formattedResponse.concat(
                lines.map((line, i) => (
                    <span key={i}>
                        {line}
                        <br />
                    </span>
                ))
            );
        }

        // Verifica se a resposta contém um link e o formata como HTML
        const linkMatch = response.match(/<a href='(.*?)' .*?>(.*?)<\/a>/);
        if (linkMatch) {
            console.log("Link detected:", linkMatch[1]);
            formattedResponse.push(
                <a href={linkMatch[1]} target="_blank" rel="noopener noreferrer" key="link">
                    {linkMatch[2]}
                </a>
            );
            response = response.replace(/<a href='(.*?)' .*?>(.*?)<\/a>/, '');
        }

        return formattedResponse;
    };

    const handleYesClick = async () => {
        console.log("User clicked YES.");
        await sendMessage("YES");
    };

    const handleNoClick = () => {
        console.log("User clicked NO.");
        setChat(prevChat => [...prevChat, { user: "You", response: "No" }, { user: "ChatApp v2", response: "I understand. What else would you like to adjust?" }]);
        setStep(1); // Volta ao step inicial se a resposta for "No"
    };

    return (
        <Container maxWidth="md" style={{ marginTop: '50px' }}>
            <Paper style={{ padding: '20px', marginBottom: '20px', backgroundColor: '#f5f5f5' }}>
                <Typography variant="h5" gutterBottom>
                    ChatApp v2
                </Typography>
                <List>
                    {chat.map((msg, index) => (
                        <ListItem key={index} style={{ textAlign: msg.user === "You" ? 'right' : 'left' }}>
                            <ListItemText 
                                primary={<strong>{msg.user}:</strong>} 
                                secondary={
                                    Array.isArray(msg.response)
                                        ? msg.response.map((item, idx) => <div key={idx}>{item}</div>)
                                        : msg.response
                                }
                                style={{ backgroundColor: msg.user === "You" ? '#e1f5fe' : '#ffffff', borderRadius: '10px', padding: '10px' }} 
                            />
                        </ListItem>
                    ))}
                </List>
                {!isChatBlocked && step !== 2 && (
                    <>
                        <TextField
                            fullWidth
                            multiline
                            rows={8}  
                            variant="outlined"
                            label="Enter your message"
                            value={message}
                            onChange={(e) => setMessage(e.target.value)}
                            style={{ marginBottom: '10px', backgroundColor: '#ffffff' }}
                        />
                        <Button
                            variant="contained"
                            color="primary"
                            onClick={() => sendMessage(message)}
                            style={{ marginTop: '10px' }}
                        >
                            Send
                        </Button>
                    </>
                )}
                {step === 2 && !isChatBlocked && (
                    <Box display="flex" justifyContent="center" mt={2}>
                        <Button
                            variant="contained"
                            color="primary"
                            onClick={handleYesClick}  
                            style={{ marginRight: '10px' }}
                        >
                            YES
                        </Button>
                        <Button
                            variant="contained"
                            color="secondary"
                            onClick={handleNoClick}  
                        >
                            NO
                        </Button>
                    </Box>
                )}
            </Paper>
        </Container>
    );    
};

export default ChatInterface;
