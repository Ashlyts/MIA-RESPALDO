// tailwind.config.js
/** @type {import('tailwindcss').Config} */
export default {
    content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
    extend: {
        colors: {
        'galaxy-bg': '#0a0f1f',
        'console-border': '#0ea5e9',
        'output-bg': '#1e293b',
        }
    },
    },
    plugins: [],
}