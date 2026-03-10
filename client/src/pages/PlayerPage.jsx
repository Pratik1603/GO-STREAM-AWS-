import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ArrowLeft } from 'lucide-react';
import { API_ENDPOINTS } from '../config';
import { useEffect, useRef } from 'react';

const PlayerPage = () => {
    const { id } = useParams();
    const { user } = useAuth();
    const navigate = useNavigate();
    const { state } = useLocation();
    const videoRef = useRef(null);
    const lastSavedTime = useRef(0);
    const saveInterval = 5; // seconds

    const resumeProgress = state?.resumeProgress || 0;

    // Construct stream URL with token
    const streamUrl = `${API_ENDPOINTS.STREAM}/${id}?token=${user?.token}`;

    const saveProgressRef = useRef(null);

    const saveProgress = async () => {
        if (!videoRef.current || !user?.token || !id) return;

        const currentTime = Math.floor(videoRef.current.currentTime);
        const duration = videoRef.current.duration;
        if (!duration) return;

        const progressPercent = Math.floor((currentTime / duration) * 100);

        // Only save if progressed at least 2 seconds or close to end, to prevent spam
        if (Math.abs(currentTime - lastSavedTime.current) < 2 && progressPercent < 95 && lastSavedTime.current !== 0) return;

        try {
            await fetch(API_ENDPOINTS.WATCH_HISTORY, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${user.token}`
                },
                body: JSON.stringify({
                    movie_id: parseInt(id),
                    progress: progressPercent
                })
            });
            lastSavedTime.current = currentTime;
        } catch (err) {
            console.error("Failed to save watch progress", err);
        }
    };

    useEffect(() => {
        saveProgressRef.current = saveProgress;
    });

    useEffect(() => {
        const handleBeforeUnload = () => {
            if (saveProgressRef.current) saveProgressRef.current();
        };

        window.addEventListener('beforeunload', handleBeforeUnload);

        return () => {
            if (saveProgressRef.current) saveProgressRef.current();
            window.removeEventListener('beforeunload', handleBeforeUnload);
        };
    }, []);

    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const handleLoadedMetadata = () => {
            if (resumeProgress > 0 && resumeProgress < 95) {
                video.currentTime = (video.duration * resumeProgress) / 100;
            }
        };

        video.addEventListener('loadedmetadata', handleLoadedMetadata);
        return () => video.removeEventListener('loadedmetadata', handleLoadedMetadata);
    }, [resumeProgress]);

    const handleTimeUpdate = () => {
        const currentTime = videoRef.current.currentTime;
        if (currentTime - lastSavedTime.current >= saveInterval) {
            saveProgress();
        }
    };

    return (
        <div className="bg-black h-screen w-full flex flex-col items-center justify-center relative group">
            <button
                onClick={() => navigate(-1)}
                className="absolute top-8 left-8 text-white hover:text-gray-300 z-50 bg-black/50 p-3 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-300 backdrop-blur-md border border-white/10"
            >
                <ArrowLeft className="w-8 h-8" />
            </button>

            <video
                ref={videoRef}
                src={streamUrl}
                controls
                autoPlay
                onTimeUpdate={handleTimeUpdate}
                onPause={saveProgress}
                className="w-full h-full object-contain"
            >
                Your browser does not support the video tag.
            </video>

            {/* Premium Overlay Hint */}
            <div className="absolute top-8 right-8 pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity duration-500">
                <span className="bg-netflix-red/80 text-white text-[10px] font-bold px-2 py-1 rounded backdrop-blur-sm">
                    STREAMING HD
                </span>
            </div>
        </div>
    );
};

export default PlayerPage;
