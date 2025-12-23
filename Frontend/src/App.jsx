// src/App.jsx
import React, { useState } from 'react';
import CommandInput from './paginas/inputcommand';
import CommandOutput from './paginas/outcomand';

function App() {
  const [outputLines, setOutputLines] = useState([]);
  const [showReports, setShowReports] = useState(false);
  const [reportFiles, setReportFiles] = useState([]);
  const [loadingReports, setLoadingReports] = useState(false);

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
        jsonData = await response.json();
      } catch (parseError) {
        throw new Error(`Error al parsear JSON: ${parseError.message}`);
      }

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

  const loadReports = async () => {
    setLoadingReports(true);
    try {
      const response = await fetch('http://localhost:9600/reportes/list');
      if (!response.ok) throw new Error('Error al cargar reportes');
      const data = await response.json();
      setReportFiles(data.data || []);
    } catch (error) {
      alert('Error al cargar la lista de reportes: ' + error.message);
      setReportFiles([]);
    } finally {
      setLoadingReports(false);
    }
  };

  const clearOutput = () => {
    setOutputLines([]);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-950 via-gray-900 to-indigo-950 text-gray-100 flex flex-col">
      {/* Header */}
      <header className="p-4 border-b border-cyan-700/30 bg-black/20 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-3xl md:text-4xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-cyan-400 via-purple-400 to-pink-400">
            GoDisk
          </h1>
          <div className="flex gap-3">
            <button
              onClick={() => {
                setShowReports(true);
                loadReports();
              }}
              className="px-4 py-2 bg-purple-700/70 hover:bg-purple-700 transition-colors rounded-lg border border-purple-500/30 text-sm"
            >
              Reportes
            </button>
            <button
              onClick={clearOutput}
              className="px-4 py-2 bg-rose-700/70 hover:bg-rose-700 transition-colors rounded-lg border border-rose-500/30 text-sm"
            >
              Limpiar
            </button>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl mx-auto w-full px-2 md:px-4 py-2 gap-4 flex">
        {/* Left Panel - Command Console */}
        <section className="w-1/2 bg-gray-800/40 backdrop-blur-md rounded-xl border border-cyan-600/30 overflow-hidden flex flex-col">
          <div className="bg-gradient-to-r from-cyan-800/40 to-cyan-900/40 px-4 py-2.5 border-b border-cyan-700/30">
            <h2 className="font-medium text-cyan-200 flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-green-400"></span>
              Consola de Comandos
            </h2>
          </div>
          <div className="flex-1 p-3 flex flex-col">
            <CommandInput onExecute={handleCommandExecution} />
          </div>
        </section>

        {/* Right Panel - Output */}
        <section className="w-1/2 bg-gray-800/40 backdrop-blur-md rounded-xl border border-purple-600/30 overflow-hidden flex flex-col">
          <div className="bg-gradient-to-r from-purple-800/40 to-purple-900/40 px-4 py-2.5 border-b border-purple-700/30 flex justify-between items-center">
            <h2 className="font-medium text-purple-200 flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-blue-400"></span>
              Salida del Sistema
            </h2>
            <span className="text-xs bg-gray-700/50 px-2 py-1 rounded">
              {outputLines.length} mensaje{outputLines.length !== 1 ? 's' : ''}
            </span>
          </div>
          <div className="flex-1 overflow-hidden">
            <CommandOutput lines={outputLines} />
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="p-3 text-center text-gray-500 text-xs border-t border-gray-800">
        <p>Sistema de Archivos EXT2 • USAC 2025</p>
      </footer>

      {/* Modal de Reportes */}
      {showReports && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-gray-800/90 border border-purple-500/50 rounded-xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
            <div className="bg-gradient-to-r from-purple-800/60 to-pink-800/60 px-6 py-4 flex justify-between items-center">
              <h2 className="text-xl font-bold text-white">Reportes Generados</h2>
              <button
                onClick={() => setShowReports(false)}
                className="text-gray-300 hover:text-white text-2xl"
              >
                &times;
              </button>
            </div>
            <div className="p-4 flex-1 overflow-y-auto">
              {loadingReports ? (
                <p className="text-center text-gray-400">Cargando...</p>
              ) : reportFiles.length === 0 ? (
                <p className="text-center text-gray-500">No hay reportes generados aún.</p>
              ) : (
                <ul className="space-y-2">
                  {reportFiles.map((file, index) => (
                    <li key={index} className="flex justify-between items-center bg-gray-900/50 p-3 rounded-lg border border-gray-700">
                      <span className="text-cyan-200 font-mono">{file}</span>
                      <div className="flex gap-2">
                        <button
                          onClick={() => window.open(`http://localhost:9600/rep/${file}`, '_blank')}
                          className="px-3 py-1 bg-cyan-700 hover:bg-cyan-600 text-xs rounded text-white"
                        >
                          Ver
                        </button>
                        <button
                          onClick={() => {
                            const link = document.createElement('a');
                            link.href = `http://localhost:9600/rep/${file}`;
                            link.download = file;
                            document.body.appendChild(link);
                            link.click();
                            document.body.removeChild(link);
                          }}
                          className="px-3 py-1 bg-purple-700 hover:bg-purple-600 text-xs rounded text-white"
                        >
                          Descargar
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <div className="px-4 py-3 bg-gray-900/50 flex justify-end">
              <button
                onClick={() => setShowReports(false)}
                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded"
              >
                Cerrar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;