import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, Bell, User, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import classNames from 'classnames';

const Navbar = () => {
    const [isScrolled, setIsScrolled] = useState(false);
    const { logout, user } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        const handleScroll = () => {
            setIsScrolled(window.scrollY > 0);
        };
        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    const handleLogout = () => {
        logout();
        navigate('/');
    };

    return (
        <nav className={classNames(
            "fixed top-0 w-full z-50 transition-colors duration-300 px-4 md:px-12 py-4 flex items-center justify-between",
            { "bg-netflix-black": isScrolled, "bg-transparent": !isScrolled }
        )}>
            <div className="flex items-center gap-8">
                <Link to="/browse" className="text-netflix-red text-2xl md:text-3xl font-bold">GO STREAM</Link>
                <div className="hidden md:flex gap-6 text-sm text-gray-200">
                    <Link to="/browse" className="hover:text-white transition">Home</Link>
                    <Link to="/browse" className="hover:text-white transition">TV Shows</Link>
                    <Link to="/browse" className="hover:text-white transition">Movies</Link>
                    <Link to="/browse" className="hover:text-white transition">New & Popular</Link>
                </div>
            </div>

            <div className="flex items-center gap-6 text-white">
                <div className="relative flex items-center">
                    <Search
                        className="w-5 h-5 cursor-pointer hover:text-gray-300 absolute left-2"
                        onClick={() => document.getElementById('search-input').focus()}
                    />
                    <input
                        id="search-input"
                        type="text"
                        placeholder="Titles, people, genres"
                        className="bg-black/40 border border-white/20 pl-9 pr-4 py-1.5 rounded text-sm w-0 focus:w-64 transition-all duration-300 focus:bg-black focus:outline-none placeholder:text-gray-500"
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                                navigate(`/browse?q=${e.target.value}`);
                            }
                        }}
                    />
                </div>
                <Bell className="w-5 h-5 cursor-pointer hover:text-gray-300" />

                <div className="group relative">
                    <div className="flex items-center gap-2 cursor-pointer">
                        <div className="w-8 h-8 rounded bg-blue-600 flex items-center justify-center">
                            <User className="w-5 h-5" />
                        </div>
                    </div>

                    <div className="absolute right-0 top-full mt-2 w-48 bg-netflix-black border border-gray-700 rounded shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200">
                        <div className="py-2">
                            <div className="px-4 py-2 border-b border-gray-700 mb-1">
                                <p className="text-sm font-bold truncate">{user?.name || user?.email}</p>
                                <p className="text-[10px] text-gray-400 uppercase tracking-wider">{user?.role || 'user'}</p>
                            </div>
                            <Link to="/profile" className="block px-4 py-2 hover:underline text-sm">Account</Link>
                            {user?.role === 'admin' && (
                                <Link to="/upload" className="block px-4 py-2 hover:underline text-netflix-red">Upload Movie</Link>
                            )}
                            <button
                                onClick={handleLogout}
                                className="w-full text-left px-4 py-2 hover:underline flex items-center gap-2"
                            >
                                <LogOut className="w-4 h-4" /> Sign out
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </nav>
    );
};

export default Navbar;
