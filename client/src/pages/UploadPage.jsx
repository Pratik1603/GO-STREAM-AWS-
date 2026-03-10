import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Upload, Film, Tv, PlaySquare } from 'lucide-react';
import { API_ENDPOINTS } from '../config';

const UploadPage = () => {
    const [contentType, setContentType] = useState('movie'); // 'movie', 'series', 'episode'
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [file, setFile] = useState(null);
    const [thumbnail, setThumbnail] = useState(null);

    // Extended Metadata
    const [parentID, setParentID] = useState('');
    const [seasonNum, setSeasonNum] = useState('');
    const [episodeNum, setEpisodeNum] = useState('');
    const [cast, setCast] = useState('');
    const [director, setDirector] = useState('');
    const [releaseYear, setReleaseYear] = useState('');
    const [genres, setGenres] = useState('');

    const [seriesList, setSeriesList] = useState([]);

    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState('');

    const { user } = useAuth();
    const navigate = useNavigate();

    // Fetch existing series for the "Episode" upload flow
    useEffect(() => {
        const fetchSeries = async () => {
            if (!user?.token) return;
            try {
                const res = await fetch(API_ENDPOINTS.MOVIES, {
                    headers: { 'Authorization': `Bearer ${user.token}` }
                });
                if (res.ok) {
                    const data = await res.json();
                    // Filter down to just series
                    const existingSeries = data.filter(m => m.content_type === 'series');
                    setSeriesList(existingSeries);
                }
            } catch (err) {
                console.error("Failed to load series for dropdown", err);
            }
        };
        fetchSeries();
    }, [user]);

    const handleUpload = async (e) => {
        e.preventDefault();
        setError('');

        if (contentType !== 'series' && !file) {
            setError("Please select a video file for this content type.");
            return;
        }
        if (contentType === 'episode' && !parentID) {
            setError("Please select a Series for this episode.");
            return;
        }

        setUploading(true);
        const formData = new FormData();
        formData.append('title', title);
        formData.append('description', description);
        formData.append('content_type', contentType);

        if (file) formData.append('video', file);
        if (thumbnail) formData.append('thumbnail', thumbnail);

        // Extended fields
        if (director) formData.append('director', director);
        if (cast) formData.append('cast_members', cast);
        if (genres) formData.append('genres', genres);
        if (releaseYear) formData.append('release_year', releaseYear);

        if (contentType === 'episode') {
            formData.append('parent_id', parentID);
            if (seasonNum) formData.append('season_number', seasonNum);
            if (episodeNum) formData.append('episode_number', episodeNum);
        }

        try {
            const res = await fetch(`${API_ENDPOINTS.ADMIN}/movies`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${user.token}`
                },
                body: formData
            });

            if (res.ok) {
                navigate('/browse');
            } else {
                const data = await res.json();
                setError(data.error || "Upload failed");
            }
        } catch (err) {
            setError("Connection error");
        } finally {
            setUploading(false);
        }
    };

    return (
        <div className="min-h-screen bg-netflix-dark pt-24 pb-20 px-4 flex justify-center">
            <div className="max-w-xl w-full bg-zinc-900 p-8 rounded-lg shadow-lg">
                <h1 className="text-3xl font-bold mb-6 flex items-center gap-3 text-white">
                    <Upload className="text-netflix-red" /> Add New Content
                </h1>

                {error && <div className="bg-orange-500 text-white p-3 rounded mb-4">{error}</div>}

                <form onSubmit={handleUpload} className="flex flex-col gap-4">
                    {/* Content Type Selector */}
                    <div>
                        <label className="block text-gray-400 mb-2 font-medium">Content Type</label>
                        <div className="grid grid-cols-3 gap-3">
                            <button
                                type="button"
                                onClick={() => setContentType('movie')}
                                className={`flex items-center justify-center gap-2 py-3 rounded-md transition font-bold ${contentType === 'movie' ? 'bg-netflix-red text-white' : 'bg-zinc-800 text-gray-400 hover:text-white'}`}
                            >
                                <Film size={18} /> Movie
                            </button>
                            <button
                                type="button"
                                onClick={() => setContentType('series')}
                                className={`flex items-center justify-center gap-2 py-3 rounded-md transition font-bold ${contentType === 'series' ? 'bg-netflix-red text-white' : 'bg-zinc-800 text-gray-400 hover:text-white'}`}
                            >
                                <Tv size={18} /> TV Series
                            </button>
                            <button
                                type="button"
                                onClick={() => setContentType('episode')}
                                className={`flex items-center justify-center gap-2 py-3 rounded-md transition font-bold ${contentType === 'episode' ? 'bg-netflix-red text-white' : 'bg-zinc-800 text-gray-400 hover:text-white'}`}
                            >
                                <PlaySquare size={18} /> Episode
                            </button>
                        </div>
                    </div>

                    {/* Conditional Episode Fields */}
                    {contentType === 'episode' && (
                        <>
                            <div>
                                <label className="block text-gray-400 mb-1">Select Series</label>
                                <select
                                    className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red"
                                    value={parentID}
                                    onChange={(e) => setParentID(e.target.value)}
                                    required
                                >
                                    <option value="" disabled>-- Choose a Series --</option>
                                    {seriesList.map(series => (
                                        <option key={series.id} value={series.id}>{series.title}</option>
                                    ))}
                                </select>
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-gray-400 mb-1">Season Number</label>
                                    <input type="number" min="1" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={seasonNum} onChange={(e) => setSeasonNum(e.target.value)} required />
                                </div>
                                <div>
                                    <label className="block text-gray-400 mb-1">Episode Number</label>
                                    <input type="number" min="1" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={episodeNum} onChange={(e) => setEpisodeNum(e.target.value)} required />
                                </div>
                            </div>
                        </>
                    )}

                    {/* Common Fields */}
                    <div>
                        <label className="block text-gray-400 mb-1">{contentType === 'episode' ? 'Episode Title' : 'Title'}</label>
                        <input type="text" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={title} onChange={(e) => setTitle(e.target.value)} required />
                    </div>

                    <div>
                        <label className="block text-gray-400 mb-1">Description</label>
                        <textarea className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red h-24" value={description} onChange={(e) => setDescription(e.target.value)} />
                    </div>

                    {/* Metadata fields (Only for Movies and Series) */}
                    {contentType !== 'episode' && (
                        <>
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-gray-400 mb-1">Release Year</label>
                                    <input type="number" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={releaseYear} onChange={(e) => setReleaseYear(e.target.value)} />
                                </div>
                                <div>
                                    <label className="block text-gray-400 mb-1">Director</label>
                                    <input type="text" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={director} onChange={(e) => setDirector(e.target.value)} />
                                </div>
                            </div>

                            <div>
                                <label className="block text-gray-400 mb-1">Genres (comma separated)</label>
                                <input type="text" placeholder="e.g. Action, Sci-Fi, Thriller" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={genres} onChange={(e) => setGenres(e.target.value)} />
                            </div>

                            <div>
                                <label className="block text-gray-400 mb-1">Cast (comma separated)</label>
                                <input type="text" placeholder="e.g. Leonardo DiCaprio, Joseph Gordon-Levitt" className="w-full bg-gray-800 p-3 rounded text-white focus:outline-none focus:ring-2 focus:ring-netflix-red" value={cast} onChange={(e) => setCast(e.target.value)} />
                            </div>
                        </>
                    )}

                    {/* Video Upload (Not for Series) */}
                    {contentType !== 'series' && (
                        <div>
                            <label className="block text-gray-400 mb-1">Video File (MP4)</label>
                            <input type="file" accept="video/mp4" className="w-full bg-gray-800 p-3 rounded text-white" onChange={(e) => setFile(e.target.files[0])} required />
                        </div>
                    )}

                    <div>
                        <label className="block text-gray-400 mb-1">{contentType === 'series' ? 'Series Poster (Image)' : 'Thumbnail (Optional)'}</label>
                        <input type="file" accept="image/*" className="w-full bg-gray-800 p-3 rounded text-white" onChange={(e) => setThumbnail(e.target.files[0])} />
                    </div>

                    <button type="submit" disabled={uploading} className="bg-netflix-red text-white py-3 my-4 rounded font-bold hover:bg-red-700 transition shadow-lg w-full text-lg disabled:opacity-50">
                        {uploading ? 'Processing & Uploading...' : `Upload ${contentType.charAt(0).toUpperCase() + contentType.slice(1)}`}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default UploadPage;
