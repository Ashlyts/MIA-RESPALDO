// src/components/ReportsPage.jsx
import React, { useState, useEffect } from 'react';

const ReportsPage = () => {
    const [selectedReport, setSelectedReport] = useState('mbr');
    const [reportId, setReportId] = useState('');
    const [reportName, setReportName] = useState('');

    const reportOptions = [
        { value: 'mbr', label: 'MBR' },
        { value: 'disk', label: 'DISK' },
        { value: 'inode', label: 'INODE' },
        { value: 'block', label: 'BLOCK' },
        { value: 'bm_inode', label: 'BM_INODE' },
        { value: 'bm_block', label: 'BM_BLOCK' },
        { value: 'tree', label: 'TREE' },
        { value: 'sb', label: 'SB' },
    ];

    const handleGenerateReport = () => {
        if (!reportId || !reportName) {
            alert('Por favor, ingrese el ID y el nombre del reporte.');
            return;
        }
        // Simulamos la generación del reporte
        alert(`Generando reporte: ${selectedReport} con ID: ${reportId} y nombre: ${reportName}`);
        // En un proyecto real, esto sería una llamada al backend.
    };

    return (
        <div className="bg-gray-800/30 backdrop-blur-sm border border-purple-500/30 rounded-xl shadow-lg shadow-indigo-950/50 p-6">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-pink-500">Generar Reportes</h1>
                <Link to="/" className="px-4 py-2 bg-cyan-700 hover:bg-cyan-800 transition-colors rounded-md border border-cyan-500/50">
                    Volver a la Consola
                </Link>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium mb-1">Tipo de Reporte</label>
                        <select
                            value={selectedReport}
                            onChange={(e) => setSelectedReport(e.target.value)}
                            className="w-full p-2 bg-gray-900/80 border border-cyan-500/30 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500/50 text-cyan-100"
                        >
                            {reportOptions.map(option => (
                                <option key={option.value} value={option.value}>
                                    {option.label}
                                </option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1">ID de la Partición</label>
                        <input
                            type="text"
                            value={reportId}
                            onChange={(e) => setReportId(e.target.value)}
                            placeholder="Ej: A118"
                            className="w-full p-2 bg-gray-900/80 border border-cyan-500/30 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500/50 text-cyan-100"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1">Nombre del Reporte</label>
                        <input
                            type="text"
                            value={reportName}
                            onChange={(e) => setReportName(e.target.value)}
                            placeholder="Ej: reporte1.jpg"
                            className="w-full p-2 bg-gray-900/80 border border-cyan-500/30 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500/50 text-cyan-100"
                        />
                    </div>
                    <button
                        onClick={handleGenerateReport}
                        className="w-full px-4 py-2 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 transition-colors rounded-md"
                    >
                        Generar Reporte
                    </button>
                </div>
                <div className="bg-gray-900/50 p-4 rounded-lg border border-gray-700">
                    <h2 className="text-lg font-semibold mb-2">Instrucciones</h2>
                    <p className="text-sm text-gray-400 mb-2">
                        Seleccione el tipo de reporte que desea generar, ingrese el ID de la partición y el nombre del archivo.
                    </p>
                    <p className="text-sm text-gray-400">
                        Los reportes disponibles son: MBR, DISK, INODE, BLOCK, BM_INODE, BM_BLOCK, TREE, SB.
                    </p>
                </div>
            </div>
        </div>
    );
};

export default ReportsPage;