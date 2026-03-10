import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const AuthPage = ({ isRegister = false }) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [name, setName] = useState('');
    const [error, setError] = useState('');
    const { login, register } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        let res;
        if (isRegister) {
            res = await register(email, password, name);
        } else {
            res = await login(email, password);
        }

        if (res.success) {
            navigate('/browse');
        } else {
            setError(res.error || 'Authentication failed');
        }
    };

    return (
        <div className="relative h-screen w-full bg-[url('https://assets.nflxext.com/ffe/siteui/vlv3/c38a2d52-138e-48a3-ab68-36787ece46b3/eeb03fc9-99bf-47d3-a6a9-a0ca713e3146/US-en-20240101-popsignuptwoweeks-perspective_alpha_website_large.jpg')] bg-cover bg-center">
            <div className="absolute inset-0 bg-black/60"></div>

            <div className="absolute top-0 w-full p-6">
                <Link to="/" className="text-netflix-red text-4xl font-bold">GO STREAM</Link>
            </div>

            <div className="relative z-10 flex justify-center items-center h-full">
                <div className="bg-black/75 p-16 rounded-lg w-full max-w-md">
                    <h2 className="text-3xl font-bold mb-8">{isRegister ? 'Sign Up' : 'Sign In'}</h2>

                    {error && <div className="bg-orange-500 text-white p-3 rounded mb-4 text-sm">{error}</div>}

                    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                        {isRegister && (
                            <input
                                type="text"
                                placeholder="Name"
                                className="bg-gray-700 text-white p-3 rounded focus:outline-none focus:bg-gray-600"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                required
                            />
                        )}
                        <input
                            type="email"
                            placeholder="Email or phone number"
                            className="bg-gray-700 text-white p-3 rounded focus:outline-none focus:bg-gray-600"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            required
                        />
                        <input
                            type="password"
                            placeholder="Password"
                            className="bg-gray-700 text-white p-3 rounded focus:outline-none focus:bg-gray-600"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                        />

                        <button type="submit" className="bg-netflix-red text-white py-3 rounded font-bold mt-6 hover:bg-red-700 transition">
                            {isRegister ? 'Sign Up' : 'Sign In'}
                        </button>
                    </form>

                    <div className="mt-4 text-gray-400">
                        {isRegister ? 'Already have an account? ' : 'New to Go Stream? '}
                        <Link to={isRegister ? '/login' : '/register'} className="text-white hover:underline">
                            {isRegister ? 'Sign in now.' : 'Sign up now.'}
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default AuthPage;
