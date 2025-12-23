// src/paginas/outcomand.jsx
import React, { useEffect, useRef } from 'react';

const CommandOutput = ({ lines }) => {
    const outputRef = useRef(null);

    useEffect(() => {
        if (outputRef.current) {
            outputRef.current.scrollTop = outputRef.current.scrollHeight;
        }
    }, [lines]);

    const getLineStyle = (type) => {
        switch (type) {
            case 'error':
                return 'text-red-400 font-medium';
            case 'success':
                return 'text-green-300';
            case 'comment':
                return 'text-yellow-400 italic';
            default:
                return 'text-cyan-100';
        }
    };

    return (
        <div
            ref={outputRef}
            className="h-full w-full overflow-y-auto p-3 font-mono text-sm leading-relaxed"
            style={{
                background:
                    'radial-gradient(circle at 1px 1px, rgba(56, 189, 248, 0.05) 1px, transparent 1px)',
                backgroundSize: '20px 20px',
            }}
        >
            {lines.length === 0 ? (
                <div className="h-full flex items-center justify-center text-gray-500">
                    <p className="text-center max-w-md">
                        Ejecuta comandos como <code className="bg-gray-800 px-2 py-1 rounded">mkdisk</code>,{' '}
                        <code className="bg-gray-800 px-2 py-1 rounded">fdisk</code>,{' '}
                        <code className="bg-gray-800 px-2 py-1 rounded">mount</code>, etc.<br />
                        <span className="text-xs mt-2 block opacity-75">
                        </span>
                    </p>
                </div>
            ) : (
                <div className="space-y-4">
                    {lines.map((line) => (
                        <div
                            key={line.id}
                            className={`${getLineStyle(line.type)} break-words whitespace-pre-wrap`}
                        >
                            {line.content}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default CommandOutput;