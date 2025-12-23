// src/paginas/inputcommand.jsx
import React, { useState, useRef } from 'react';

const CommandInput = ({ onExecute }) => {
    const [inputValue, setInputValue] = useState('');
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

    const handleKeyDown = (e) => {
        if (e.ctrlKey && e.key === 'Enter') {
            e.preventDefault();
            handleExecute();
        }
    };

    return (
        <div className="flex flex-col h-full">
            {/* Botones */}
            <div className="flex flex-wrap gap-2 mb-3">
                <label className="cursor-pointer bg-cyan-800/60 hover:bg-cyan-700 transition-colors px-3 py-1.5 rounded text-sm border border-cyan-600/40 flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                    Cargar Script
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
                    className={`px-4 py-1.5 rounded text-sm font-medium transition-all flex items-center gap-1.5 ${inputValue.trim()
                        ? 'bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 shadow shadow-purple-500/20'
                        : 'bg-gray-700/50 cursor-not-allowed text-gray-500'
                        }`}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clipRule="evenodd" />
                    </svg>
                    Ejecutar
                </button>
            </div>
            <div className="relative flex-1 overflow-y-auto rounded-lg border border-cyan-600/30 bg-gray-900/80 p-3">
                <textarea
                    ref={textareaRef}
                    value={inputValue}
                    onChange={(e) => setInputValue(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder={`Ejemplos:\n• mkdisk -size=10 -unit=M -name=disco1.mia\n• fdisk -size=5 -unit=M -diskname=disco1.mia -name=part1 -type=P\n• mount -diskname=disco1.mia -name=part1`}
                    className="w-full h-full p-0 bg-transparent border-none outline-none font-mono text-cyan-50 resize-none"
                    style={{
                        // Estilo para scroll personalizado
                        scrollbarWidth: 'thin',
                        scrollbarColor: '#3b82f6 #1f2937', // track + thumb
                        minHeight: '120px',
                    }}
                />
                {inputValue && (
                    <button
                        onClick={() => setInputValue('')}
                        className="absolute top-2 right-2 text-gray-400 hover:text-white p-1 z-10"
                        title="Limpiar"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                        </svg>
                    </button>
                )}
            </div>
        </div>
    );
};

export default CommandInput;