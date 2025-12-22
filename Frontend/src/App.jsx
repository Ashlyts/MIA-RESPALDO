// src/App.jsx
import React, { useState } from 'react';
import CommandInput from './paginas/inputcommand'; // Asegúrate que la ruta sea correcta
import CommandOutput from './paginas/outcomand'; // Asegúrate que la ruta sea correcta

function App() {
  const [outputLines, setOutputLines] = useState([]);

  const handleCommandExecution = async (commandText) => {
    const lines = commandText.split('\n').map(line => line.trim()).filter(line => line);

    try {
      const response = await fetch('http://localhost:9600/commands', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ Comandos: commandText }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      let jsonData;
      try {
        jsonData = await response.json(); // ✅ Parsea el JSON
      } catch (parseError) {
        throw new Error(`Error al parsear JSON: ${parseError.message}`);
      }

      // Extraer el contenido de 'data' (array de strings)
      const outputLines = jsonData.data || [];
      const outputText = outputLines.join('\n') || 'Comando ejecutado sin salida.';

      const resultLine = {
        id: Date.now(),
        type: jsonData.error ? 'error' : 'success',
        content: outputText
      };

      setOutputLines(prev => [...prev, resultLine]);
    } catch (error) {
      const errorLine = {
        id: Date.now() + lines.length + 1,
        type: 'error',
        content: `Error al conectar con el backend: ${error.message}`
      };
      setOutputLines(prev => [...prev, errorLine]);
    }
  };

  const clearOutput = () => {
    setOutputLines([]);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-black to-indigo-950 text-gray-100 p-4 md:p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <header className="mb-8 flex justify-between items-center">
          <h1 className="text-4xl md:text-6xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-cyan-400 via-purple-500 to-pink-500">
            GoDisk
          </h1>
          <div className="flex gap-2">
            {/* Botón de reportes temporal, puedes ocultarlo o dejarlo como placeholder */}
            <button
              onClick={() => alert("Funcionalidad de reportes en construcción. Necesitas react-router-dom para la navegación completa.")}
              className="px-4 py-2 bg-purple-700 hover:bg-purple-800 transition-colors rounded-md border border-purple-500/50"
            >
              Reportes
            </button>
            <button
              onClick={clearOutput}
              className="px-4 py-2 bg-rose-700 hover:bg-rose-800 transition-colors rounded-md border border-rose-500/50"
            >
              Limpiar Área
            </button>
          </div>
        </header>

        {/* Main Content */}
        <main className="space-y-8">
          {/* Input Section */}
          <section className="bg-gray-800/30 backdrop-blur-sm border border-cyan-500/30 rounded-xl shadow-lg shadow-indigo-950/50 overflow-hidden">
            <div className="bg-gradient-to-r from-cyan-700/50 to-purple-700/50 px-4 py-2 border-b border-cyan-500/30">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <span className="inline-block w-3 h-3 rounded-full bg-red-500 animate-pulse"></span>
                Consola de Comandos (Nuevo para Escribir)
              </h2>
            </div>
            <div className="p-4">
              <CommandInput onExecute={handleCommandExecution} />
            </div>
          </section>

          {/* Output Section */}
          <section className="bg-gray-800/30 backdrop-blur-sm border border-purple-500/30 rounded-xl shadow-lg shadow-indigo-950/50 overflow-hidden">
            <div className="bg-gradient-to-r from-purple-700/50 to-pink-700/50 px-4 py-2 border-b border-purple-500/30">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <span className="inline-block w-3 h-3 rounded-full bg-green-500"></span>
                Salida de Comandos & Visualización de Documentos
              </h2>
            </div>
            <div className="p-4 max-h-[50vh] overflow-y-auto">
              <CommandOutput lines={outputLines} />
            </div>
          </section>
        </main>

        {/* Footer */}
        <footer className="mt-12 text-center text-gray-400 text-sm">
          <p>Sistema de Archivos EXT2 - Universidad San Carlos de Guatemala</p>
        </footer>
      </div>
    </div>
  );
}

export default App;