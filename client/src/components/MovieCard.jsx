import { Play, RotateCcw } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { API_ENDPOINTS } from '../config';
import { useAuth } from '../context/AuthContext';

const MovieCard = ({ movie, progress = 0 }) => {
    const navigate = useNavigate();
    const { user } = useAuth();

    const handleCardClick = () => {
        navigate(`/title/${movie.id}`);
    };

    const handleResume = (e) => {
        e.stopPropagation();
        navigate(`/watch/${movie.id}`, { state: { resumeProgress: progress } });
    };

    const handleRestart = (e) => {
        e.stopPropagation();
        navigate(`/watch/${movie.id}`, { state: { resumeProgress: 0 } });
    };

    return (
        <div
            className="relative group bg-zinc-900 rounded-md overflow-hidden cursor-pointer transition-transform duration-300 hover:scale-105 hover:z-10 w-full aspect-video"
            onClick={handleCardClick}
        >
            {/* Thumbnail or Placeholder */}
            <img
                src={movie.thumbnail_url || (movie.thumbnail_path ? `${API_ENDPOINTS.MOVIES}/${movie.id}/thumbnail?token=${user?.token}` : `https://via.placeholder.com/320x180/000000/FFFFFF/?text=${encodeURIComponent(movie.title)}`)}
                alt={movie.title}
                className="w-full h-full object-cover"
            />

            {/* Hover Overlay */}
            <div className="absolute inset-0 bg-black/70 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-center items-center p-4">
                {progress > 0 ? (
                    <div className="flex flex-col gap-2 mb-2 w-full max-w-[160px]">
                        <button
                            onClick={handleResume}
                            className="bg-netflix-red text-white text-xs font-bold py-1.5 rounded flex items-center justify-center gap-1.5 hover:bg-red-700 transition"
                        >
                            <Play fill="white" className="w-3.5 h-3.5" /> Resume
                        </button>
                        <button
                            onClick={handleRestart}
                            className="bg-zinc-700 text-white text-xs font-bold py-1.5 rounded flex items-center justify-center gap-1.5 hover:bg-zinc-600 transition"
                        >
                            <RotateCcw className="w-3.5 h-3.5" /> Restart
                        </button>
                    </div>
                ) : (
                    <div className="bg-white rounded-full p-3 mb-3 text-black hover:bg-gray-200 transition">
                        <Play fill="black" className="w-5 h-5 ml-0.5" />
                    </div>
                )}
                <h3 className="text-white font-bold text-center text-sm md:text-base">{movie.title}</h3>
                <p className="text-gray-300 text-xs mt-1 line-clamp-2">{movie.description}</p>
            </div>
        </div>
    );
};

export default MovieCard;
