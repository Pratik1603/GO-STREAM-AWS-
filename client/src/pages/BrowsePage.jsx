import { useState, useEffect } from 'react';
import { API_ENDPOINTS } from '../config';
import { useAuth } from '../context/AuthContext';
import classNames from 'classnames';
import { Play, Info } from 'lucide-react';
import { useNavigate, useLocation } from 'react-router-dom';
import MovieCard from '../components/MovieCard';

const BrowsePage = () => {
    const [movies, setMovies] = useState([]);
    const [recommended, setRecommended] = useState([]);
    const [history, setHistory] = useState([]);
    const [featured, setFeatured] = useState(null);
    const [loading, setLoading] = useState(true);
    const { user } = useAuth();
    const navigate = useNavigate();

    const location = useLocation();
    const queryParams = new URLSearchParams(location.search);
    const searchQuery = queryParams.get('q');

    useEffect(() => {
        if (searchQuery) {
            fetchSearchResults();
        } else {
            fetchBrowseData();
        }
    }, [user, searchQuery]);

    const fetchBrowseData = async () => {
        if (!user?.token) return;
        setLoading(true);

        try {
            const res = await fetch(API_ENDPOINTS.BROWSE, {
                headers: { 'Authorization': `Bearer ${user.token}` }
            });
            const data = await res.json();

            if (data.all_movies) setMovies(data.all_movies);
            if (data.recommendations) setRecommended(data.recommendations);
            if (data.watch_history) setHistory(data.watch_history);

            if (data.all_movies && data.all_movies.length > 0) {
                const random = data.all_movies[Math.floor(Math.random() * data.all_movies.length)];
                setFeatured(random);
            }
        } catch (err) {
            console.error("Failed to fetch browse data", err);
        } finally {
            setLoading(false);
        }
    };

    const fetchSearchResults = async () => {
        if (!user?.token) return;
        setLoading(true);

        try {
            const res = await fetch(`${API_ENDPOINTS.MOVIES}/search?q=${encodeURIComponent(searchQuery)}`, {
                headers: { 'Authorization': `Bearer ${user.token}` }
            });
            const data = await res.json();
            if (Array.isArray(data)) {
                setMovies(data);
                setFeatured(null);
            }
        } catch (err) {
            console.error("Failed to fetch search results", err);
        } finally {
            setLoading(false);
        }
    };

    const MovieRow = ({ title, items, isHistory = false }) => {
        if (!items || items.length === 0) return null;

        return (
            <div className="mb-12">
                <h2 className="text-xl md:text-2xl font-semibold mb-4 text-white hover:text-gray-300 transition cursor-default px-4 md:px-12 flex items-center gap-2">
                    {title}
                    <Info className="w-4 h-4 opacity-50" />
                </h2>
                <div className="relative group">
                    <div className="flex overflow-x-auto gap-4 px-4 md:px-12 pb-6 no-scrollbar scroll-smooth">
                        {items.map(item => {
                            const movie = isHistory ? item.movie : item;
                            if (!movie) return null;
                            const progress = isHistory ? item.progress : 0;
                            return (
                                <div key={isHistory ? `${movie.id}-${item.id}` : movie.id} className="flex-none w-[280px] md:w-[320px]">
                                    <MovieCard movie={movie} progress={progress} />
                                    {isHistory && progress > 0 && (
                                        <div className="mt-2 px-1">
                                            <div className="h-1 w-full bg-zinc-800 rounded-full overflow-hidden">
                                                <div
                                                    className="h-full bg-netflix-red"
                                                    style={{ width: `${Math.min(progress, 100)}%` }}
                                                ></div>
                                            </div>
                                            <p className="text-[10px] text-gray-400 mt-1 uppercase tracking-wider font-semibold">
                                                Resume {movie.title}
                                            </p>
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        );
    };

    if (loading && !featured && movies.length === 0) {
        return (
            <div className="h-screen flex items-center justify-center bg-netflix-dark">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-netflix-red"></div>
            </div>
        );
    }

    return (
        <div className="pb-20 text-white selection:bg-netflix-red selection:text-white">
            {/* Hero Section */}
            {!searchQuery && featured && (
                <div className="relative h-[85vh] w-full mb-10 group overflow-hidden">
                    <div className="absolute inset-0">
                        <img
                            src={featured.thumbnail_path ? `${API_ENDPOINTS.MOVIES}/${featured.id}/thumbnail?token=${user?.token}` : `https://via.placeholder.com/1920x1080/000000/FFFFFF/?text=${encodeURIComponent(featured.title)}`}
                            alt={featured.title}
                            className="w-full h-full object-cover brightness-[0.6] transition-transform duration-700 group-hover:scale-105"
                        />
                    </div>
                    {/* Multi-layered Gradients for Premium Look */}
                    <div className="absolute inset-0 bg-gradient-to-t from-netflix-dark via-netflix-dark/40 to-transparent"></div>
                    <div className="absolute inset-0 bg-gradient-to-r from-netflix-dark via-netflix-dark/20 to-transparent"></div>
                    <div className="absolute inset-0 bg-gradient-to-l from-transparent via-transparent to-netflix-dark/10"></div>

                    <div className="absolute bottom-[25%] left-4 md:left-12 max-w-2xl px-4 animate-in fade-in slide-in-from-bottom-5 duration-700">
                        <div className="flex items-center gap-2 mb-4">
                            <span className="bg-netflix-red text-white text-[10px] font-bold px-1.5 py-0.5 rounded shadow-sm">AI RECOMMENDED</span>
                        </div>
                        <h1 className="text-4xl md:text-7xl font-extrabold mb-4 tracking-tight drop-shadow-2xl">{featured.title}</h1>
                        <p className="text-base md:text-lg text-gray-200 mb-8 line-clamp-3 max-w-lg drop-shadow-lg leading-relaxed">{featured.description}</p>

                        <div className="flex flex-wrap gap-4">
                            <button
                                onClick={() => navigate(`/watch/${featured.id}`)}
                                className="bg-white text-black px-10 py-3.5 rounded-md flex items-center gap-3 font-bold hover:bg-white/90 active:scale-95 transition-all shadow-xl"
                            >
                                <Play fill="black" className="w-6 h-6" /> Play
                            </button>
                            <button
                                onClick={() => navigate(`/title/${featured.id}`)}
                                className="bg-zinc-500/50 text-white px-10 py-3.5 rounded-md flex items-center gap-3 font-bold hover:bg-zinc-500/30 active:scale-95 transition-all backdrop-blur-md border border-white/10 shadow-xl"
                            >
                                <Info className="w-6 h-6 border-2 border-current rounded-full p-0.5" /> More Info
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Browse Sections */}
            <div className={classNames("relative z-20 transition-all duration-500", { "mt-[-15vh]": !searchQuery && featured, "pt-24": searchQuery || !featured })}>
                {searchQuery ? (
                    <div className="px-4 md:px-12">
                        <h2 className="text-2xl md:text-3xl font-bold mb-8 text-white tracking-tight">
                            Results for <span className="text-netflix-red">"{searchQuery}"</span>
                        </h2>
                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5 gap-6">
                            {movies.map(movie => (
                                <MovieCard key={movie.id} movie={movie} />
                            ))}
                        </div>
                        {movies.length === 0 && (
                            <div className="text-center py-32">
                                <Info className="w-16 h-16 mx-auto mb-6 text-zinc-700" />
                                <p className="text-zinc-500 text-xl">We couldn't find any matches for "{searchQuery}"</p>
                                <button onClick={() => navigate('/browse')} className="mt-8 text-netflix-red hover:underline font-semibold">Clear search</button>
                            </div>
                        )}
                    </div>
                ) : (
                    <>
                        <MovieRow title="Recently Watched" items={history} isHistory={true} />
                        <MovieRow title="Recommended for You" items={recommended} />

                        {/* Featured Category - Example "Trending" */}
                        <div className="mb-12">
                            <h2 className="text-xl md:text-2xl font-semibold mb-4 text-white px-4 md:px-12">Trending Now</h2>
                            <div className="flex overflow-x-auto gap-4 px-4 md:px-12 pb-6 no-scrollbar">
                                {movies.length > 0 ? (
                                    movies.slice(0, 10).map(movie => (
                                        <div key={movie.id} className="flex-none w-[280px] md:w-[320px]">
                                            <MovieCard movie={movie} />
                                        </div>
                                    ))
                                ) : (
                                    <div className="w-full text-center py-10 text-zinc-600 border border-zinc-900 rounded-xl mx-4 md:mx-12 dashed">
                                        No trending content yet
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* All Content Grid */}
                        <div className="px-4 md:px-12 pb-20 mt-20">
                            <div className="flex items-center justify-between mb-8">
                                <h2 className="text-xl md:text-2xl font-bold text-white tracking-tight">Explore Library</h2>
                                <span className="text-zinc-500 text-sm font-medium">{movies.length} titles</span>
                            </div>
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5 gap-6">
                                {movies.map(movie => (
                                    <MovieCard key={movie.id} movie={movie} />
                                ))}
                            </div>
                        </div>
                    </>
                )}
            </div>

            {/* SEO and Footer metadata hints */}
            <div className="hidden">
                <h1>StreamApp - Premium AI Recommendations</h1>
                <p>Enjoy personalized movie recommendations and seamless streaming.</p>
            </div>
        </div>
    );
};

export default BrowsePage;
