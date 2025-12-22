// src/components/CommandInput.jsx
import React, { useState, useRef } from 'react';

const CommandInput = ({ onExecute }) => {
    const [inputValue, setInputValue] = useState('');
    const [singleCommand, setSingleCommand] = useState('');
    const textareaRef = useRef(null);

    const handleFileUpload = (event) => {
        const file = event.target.files[0];
        if (file && file.name.endsWith('.mia')) {
            const reader = new FileReader();
            reader.onload = (e) => {
                const content = e.target.result;
                setInputValue(content);
            };
            reader.readAsText(file);
        } else {
            alert('Por favor, selecciona un archivo .mia válido.');
        }
    };

    const handleExecute = () => {
        if (inputValue.trim()) {
            onExecute(inputValue);
        }
    };

    const handleExecuteSingle = () => {
        if (singleCommand.trim()) {
            onExecute(singleCommand);
            setSingleCommand(''); // Limpiar el campo después de ejecutar
        }
    };

    const handleKeyDown = (e) => {
        if (e.ctrlKey && e.key === 'Enter') {
            e.preventDefault();
            handleExecute();
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex flex-col sm:flex-row gap-2">
                <label className="cursor-pointer bg-cyan-700/50 hover:bg-cyan-700 transition-colors px-4 py-2 rounded-md text-center border border-cyan-500/50 flex items-center justify-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                    Cargar Script (.mia)
                    <input
                        type="file"
                        accept=".mia"
                        onChange={handleFileUpload}
                        className="hidden"
                    />
                </label>
                <button
                    onClick={handleExecute}
                    disabled={!inputValue.trim()}
                    className={`px-6 py-2 rounded-md font-medium transition-all ${inputValue.trim()
                            ? 'bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 shadow-lg shadow-purple-500/20'
                            : 'bg-gray-700 cursor-not-allowed'
                        }`}
                >
                    Ejecutar (Ctrl+Enter)
                </button>
            </div>
            <div className="relative">
                <textarea
                    ref={textareaRef}
                    value={inputValue}
                    onChange={(e) => setInputValue(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Ingresa tus comandos aquí (ej: mkdisk -size=10 -unit=M) o carga un script..."
                    className="w-full h-40 p-4 bg-gray-900/80 border border-cyan-500/30 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500/50 resize-none font-mono text-cyan-100"
                />
                {inputValue && (
                    <button
                        onClick={() => setInputValue('')}
                        className="absolute top-2 right-2 text-gray-400 hover:text-white"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                        </svg>
                    </button>
                )}
            </div>
        </div>
    );
};

export default CommandInput;