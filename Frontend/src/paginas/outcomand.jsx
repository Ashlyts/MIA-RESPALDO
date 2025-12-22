// src/components/CommandOutput.jsx
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
                return 'text-red-400';
            case 'success':
                return 'text-green-400';
            case 'comment':
                return 'text-yellow-400 italic';
            case 'info':
            default:
                return 'text-cyan-200';
        }
    };

    return (
        <div
            ref={outputRef}
            className="font-mono text-sm space-y-1 h-full overflow-y-auto bg-gray-900/50 p-4 rounded-lg border border-gray-700"
        >
            {lines.length === 0 ? (
                <p className="text-gray-500 italic">Esperando comandos...</p>
            ) : (
                lines.map((line) => (
                    <div key={line.id} className={`${getLineStyle(line.type)} break-words whitespace-pre-wrap`}>
                        {line.content}
                    </div>
                ))
            )}
        </div>
    );
};

export default CommandOutput;