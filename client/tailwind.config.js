/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                netflix: {
                    red: '#E50914',
                    black: '#141414',
                    dark: '#141414',
                    gray: '#333333',
                    light: '#e5e5e5'
                }
            }
        },
    },
    plugins: [],
}
