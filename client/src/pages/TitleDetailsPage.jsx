import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { API_ENDPOINTS } from '../config';
import { Play, ArrowLeft, Info, Calendar, Users, Film, Trash2, Pencil, X, Check, Loader2 } from 'lucide-react';

const TitleDetailsPage = () => {
    const { id } = useParams();
    const [titleData, setTitleData] = useState(null);
    const [episodes, setEpisodes] = useState([]);
    const [similarMovies, setSimilarMovies] = useState([]);
    const [loading, setLoading] = useState(true);
    const [loadingSimilar, setLoadingSimilar] = useState(false);
    const [error, setError] = useState('');

    // Admin edit state
    const [editing, setEditing] = useState(false);
    const [saving, setSaving] = useState(false);
    const [editForm, setEditForm] = useState({});
    const [deleteConfirm, setDeleteConfirm] = useState(false);
    const [deleting, setDeleting] = useState(false);

    const { user } = useAuth();
    const navigate = useNavigate();
    // The /api/me response returns { id, role } - the field is "role"
    const isAdmin = user?.role === 'admin';

    // Debug: log user info once on mount to confirm role detection
    useEffect(() => {
        if (user) console.log('[TitleDetails] user role:', user.role, '| isAdmin:', user?.role === 'admin');
    }, [user]);

    useEffect(() => {
        const fetchWithRetry = async (url, options, retries = 3) => {
            for (let attempt = 0; attempt < retries; attempt++) {
                try {
                    const res = await fetch(url, options);
                    if (res.ok) return res;
                    if (res.status === 404 || res.status === 403) return res;
                    if (attempt < retries - 1) {
                        await new Promise(r => setTimeout(r, 500 * (attempt + 1)));
                        continue;
                    }
                    return res;
                } catch (err) {
                    if (attempt < retries - 1) {
                        await new Promise(r => setTimeout(r, 500 * (attempt + 1)));
                        continue;
                    }
                    throw err;
                }
            }
        };

        const fetchDetails = async () => {
            if (!user?.token) return;
            setLoading(true);
            setError('');
            try {
                const res = await fetchWithRetry(
                    `${API_ENDPOINTS.MOVIES}/${id}`,
                    { headers: { 'Authorization': `Bearer ${user.token}` } }
                );

                if (res && res.ok) {
                    const data = await res.json();
                    // Handle both shapes: flat movie OR {movie, episodes} for series
                    if (data.movie) {
                        setTitleData(data.movie);
                        setEpisodes(data.episodes || []);
                    } else {
                        setTitleData(data);
                        setEpisodes([]);
                    }
                } else {
                    const status = res ? res.status : 'unknown';
                    setError(status === 404 ? 'Title not found' : 'Failed to load title details');
                }
            } catch (err) {
                setError('Connection error – please try again');
            } finally {
                setLoading(false);
            }
        };

        fetchDetails();
    }, [id, user]);

    // Fetch similar movies
    useEffect(() => {
        const fetchSimilarMovies = async () => {
            if (!user?.token || !id) return;
            setLoadingSimilar(true);
            try {
                const res = await fetch(
                    `${API_ENDPOINTS.MOVIES}/${id}/similar?limit=8`,
                    { headers: { 'Authorization': `Bearer ${user.token}` } }
                );
                if (res.ok) {
                    const data = await res.json();
                    setSimilarMovies(data.similar || []);
                }
            } catch (err) {
                console.error('Failed to fetch similar movies:', err);
            } finally {
                setLoadingSimilar(false);
            }
        };

        fetchSimilarMovies();
    }, [id, user]);

    const handleEditOpen = () => {
        setEditForm({
            title: titleData.title || '',
            description: titleData.description || '',
            director: titleData.director || '',
            release_year: titleData.release_year || '',
            genres: (titleData.genres || []).join(', '),
            cast_members: (titleData.cast_members || []).join(', '),
        });
        setEditing(true);
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            const body = {
                title: editForm.title,
                description: editForm.description,
                director: editForm.director,
                release_year: parseInt(editForm.release_year) || 0,
                genres: editForm.genres.split(',').map(s => s.trim()).filter(Boolean),
                cast_members: editForm.cast_members.split(',').map(s => s.trim()).filter(Boolean),
            };
            const res = await fetch(`${API_ENDPOINTS.MOVIES}/${id}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${user.token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(body),
            });
            if (res.ok) {
                setTitleData(prev => ({
                    ...prev,
                    ...body,
                    genres: body.genres,
                    cast_members: body.cast_members,
                }));
                setEditing(false);
            } else {
                alert('Failed to save changes');
            }
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = async () => {
        setDeleting(true);
        try {
            const res = await fetch(`${API_ENDPOINTS.ADMIN}/movies/${id}`, {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${user.token}` },
            });
            if (res.ok) {
                navigate('/browse');
            } else {
                alert('Failed to delete');
                setDeleteConfirm(false);
            }
        } finally {
            setDeleting(false);
        }
    };

    const handleDeleteEpisode = async (epId) => {
        if (!window.confirm('Delete this episode?')) return;
        const res = await fetch(`${API_ENDPOINTS.ADMIN}/movies/${epId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${user.token}` },
        });
        if (res.ok) {
            setEpisodes(prev => prev.filter(e => e.id !== epId));
        } else {
            alert('Failed to delete episode');
        }
    };

    // ------------- Render Loading / Error -------------
    if (loading) {
        return (
            <div className="min-h-screen bg-netflix-dark flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-netflix-red"></div>
            </div>
        );
    }

    if (error || !titleData) {
        return (
            <div className="min-h-screen bg-netflix-dark flex flex-col items-center justify-center gap-4">
                <p className="text-white text-xl">{error || 'Title not found'}</p>
                <button onClick={() => navigate('/browse')} className="text-netflix-red flex items-center gap-2 hover:underline">
                    <ArrowLeft size={20} /> Back to Browse
                </button>
            </div>
        );
    }

    const isSeries = titleData.content_type === 'series';
    const hasVideo = titleData.video_url || titleData.file_path;
    const displayThumbnail = titleData.thumbnail_url ||
        (titleData.thumbnail_path
            ? `${API_ENDPOINTS.MOVIES}/${titleData.id}/thumbnail`
            : `https://placehold.co/1920x1080/111/333?text=${encodeURIComponent(titleData.title)}`);

    // ------------- Edit Modal -------------
    if (editing) {
        const field = (label, key, type = 'text', placeholder = '') => (
            <div>
                <label className="block text-gray-400 text-sm mb-1">{label}</label>
                {key === 'description'
                    ? <textarea rows={3} className="w-full bg-zinc-800 rounded p-2 text-white text-sm focus:outline-none focus:ring-1 focus:ring-netflix-red" value={editForm[key]} onChange={e => setEditForm(prev => ({ ...prev, [key]: e.target.value }))} />
                    : <input type={type} className="w-full bg-zinc-800 rounded p-2 text-white text-sm focus:outline-none focus:ring-1 focus:ring-netflix-red" value={editForm[key]} placeholder={placeholder} onChange={e => setEditForm(prev => ({ ...prev, [key]: e.target.value }))} />
                }
            </div>
        );

        return (
            <div className="min-h-screen bg-netflix-dark flex items-start justify-center pt-24 pb-20 px-4">
                <div className="w-full max-w-xl bg-zinc-900 rounded-xl p-8 shadow-2xl">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-2xl font-bold text-white">Edit Title</h2>
                        <button onClick={() => setEditing(false)} className="text-gray-400 hover:text-white"><X size={22} /></button>
                    </div>
                    <div className="flex flex-col gap-4">
                        {field('Title', 'title')}
                        {field('Description', 'description')}
                        {field('Director', 'director')}
                        {field('Release Year', 'release_year', 'number')}
                        {field('Genres (comma separated)', 'genres', 'text', 'e.g. Action, Sci-Fi')}
                        {field('Cast (comma separated)', 'cast_members', 'text', 'e.g. Actor A, Actor B')}
                        <div className="flex gap-3 mt-2">
                            <button onClick={handleSave} disabled={saving} className="flex-1 bg-netflix-red text-white py-2.5 rounded font-bold flex items-center justify-center gap-2 hover:bg-red-700 transition disabled:opacity-50">
                                {saving ? <Loader2 size={18} className="animate-spin" /> : <Check size={18} />} Save Changes
                            </button>
                            <button onClick={() => setEditing(false)} className="flex-1 bg-zinc-700 text-white py-2.5 rounded font-bold hover:bg-zinc-600 transition">
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    // ------------- Main Render -------------
    return (
        <div className="min-h-screen bg-netflix-dark text-white pb-20">
            {/* Hero Backdrop */}
            <div className="relative h-[60vh] md:h-[70vh] w-full">
                <div className="absolute inset-0">
                    <img src={displayThumbnail} alt={titleData.title} className="w-full h-full object-cover brightness-[0.4]" />
                </div>
                <div className="absolute inset-0 bg-gradient-to-t from-netflix-dark via-netflix-dark/60 to-transparent"></div>

                {/* Back button — offset from top so it's below the Navbar */}
                <div className="absolute top-20 left-4 md:left-12 z-20">
                    <button onClick={() => navigate('/browse')} className="flex items-center gap-2 text-white/80 hover:text-white transition bg-black/40 px-4 py-2 rounded-full backdrop-blur-sm">
                        <ArrowLeft size={20} /> Back
                    </button>
                </div>

                <div className="absolute bottom-0 left-0 w-full p-4 md:p-12 z-10">
                    <div className="max-w-4xl">
                        <div className="flex flex-wrap items-center gap-3 mb-4 text-xs md:text-sm font-semibold">
                            <span className="bg-netflix-red text-white px-2 py-1 rounded tracking-wider uppercase">
                                {titleData.content_type || 'Movie'}
                            </span>
                            {titleData.release_year > 0 && (
                                <span className="flex items-center gap-1.5 text-gray-300"><Calendar size={14} /> {titleData.release_year}</span>
                            )}
                            {titleData.duration > 0 && (
                                <span className="text-gray-300">{Math.floor(titleData.duration / 60)}h {titleData.duration % 60}m</span>
                            )}
                        </div>

                        <h1 className="text-4xl md:text-6xl font-extrabold mb-4 tracking-tight drop-shadow-2xl">{titleData.title}</h1>

                        {!isSeries && hasVideo && (
                            <button onClick={() => navigate(`/watch/${titleData.id}`)} className="mb-6 bg-white text-black px-8 py-3 rounded flex items-center gap-3 font-bold hover:bg-white/90 active:scale-95 transition-transform">
                                <Play fill="black" size={20} /> Play
                            </button>
                        )}

                        <p className="text-base md:text-lg text-gray-200 line-clamp-4 max-w-3xl leading-relaxed">
                            {titleData.description || 'No description available.'}
                        </p>
                    </div>
                </div>
            </div>

            {/* Admin Control Bar — rendered below the hero, always accessible */}
            {isAdmin && (
                <div className="flex items-center justify-end gap-3 px-4 md:px-12 py-3 bg-zinc-900/80 border-b border-zinc-800 sticky top-16 z-30 backdrop-blur-sm">
                    <span className="text-xs text-zinc-500 mr-auto uppercase tracking-wider font-semibold">Admin Controls</span>
                    <button
                        onClick={handleEditOpen}
                        className="flex items-center gap-1.5 bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded text-sm font-bold transition"
                    >
                        <Pencil size={14} /> Edit
                    </button>
                    {!deleteConfirm ? (
                        <button
                            onClick={() => setDeleteConfirm(true)}
                            className="flex items-center gap-1.5 bg-red-700 hover:bg-red-600 text-white px-4 py-2 rounded text-sm font-bold transition"
                        >
                            <Trash2 size={14} /> Delete
                        </button>
                    ) : (
                        <div className="flex items-center gap-3 bg-zinc-800 px-4 py-2 rounded border border-red-800">
                            <span className="text-sm text-white">Delete this title?</span>
                            <button
                                onClick={handleDelete}
                                disabled={deleting}
                                className="text-sm font-bold bg-red-600 hover:bg-red-500 text-white px-3 py-1 rounded transition disabled:opacity-50"
                            >
                                {deleting ? 'Deleting...' : 'Yes, Delete'}
                            </button>
                            <button
                                onClick={() => setDeleteConfirm(false)}
                                className="text-sm text-gray-400 hover:text-white"
                            >
                                Cancel
                            </button>
                        </div>
                    )}
                </div>
            )}
            <div className="px-4 md:px-12 py-8 max-w-7xl mx-auto flex flex-col lg:flex-row gap-12">
                {/* Sidebar Metadata */}
                <div className="lg:w-1/3 flex flex-col gap-5 text-sm">
                    {titleData.genres?.length > 0 && (
                        <div>
                            <span className="text-gray-400 block mb-2 font-medium">Genres</span>
                            <div className="flex flex-wrap gap-2">
                                {titleData.genres.map(g => (
                                    <span key={g} className="bg-zinc-800 px-3 py-1 rounded-full text-xs font-semibold text-white border border-zinc-700">{g}</span>
                                ))}
                            </div>
                        </div>
                    )}
                    {titleData.cast_members?.length > 0 && (
                        <div>
                            <span className="text-gray-400 block mb-1 font-medium flex items-center gap-2"><Users size={14} /> Cast</span>
                            <span className="text-white leading-relaxed">{titleData.cast_members.join(', ')}</span>
                        </div>
                    )}
                    {titleData.director && (
                        <div>
                            <span className="text-gray-400 block mb-1 font-medium flex items-center gap-2"><Film size={14} /> Director</span>
                            <span className="text-white">{titleData.director}</span>
                        </div>
                    )}
                </div>

                {/* Episode List for Series */}
                <div className="lg:w-2/3">
                    {isSeries && (
                        <div>
                            <h2 className="text-2xl font-bold mb-6 border-b border-zinc-800 pb-2">Episodes</h2>
                            {episodes.length > 0 ? (
                                <div className="flex flex-col gap-4">
                                    {episodes.map(ep => (
                                        <div key={ep.id} className="group flex flex-col sm:flex-row gap-4 p-4 rounded-lg bg-zinc-900/50 hover:bg-zinc-800 transition border border-transparent hover:border-zinc-700">
                                            <div
                                                onClick={() => navigate(`/watch/${ep.id}`)}
                                                className="relative w-full sm:w-48 aspect-video flex-shrink-0 bg-black rounded overflow-hidden cursor-pointer"
                                            >
                                                <img
                                                    src={ep.thumbnail_url || `https://placehold.co/320x180/111/333?text=Ep+${ep.episode_number}`}
                                                    alt={ep.title}
                                                    className="w-full h-full object-cover transition duration-300 group-hover:brightness-75"
                                                />
                                                <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                                    <div className="bg-black/60 rounded-full p-2 border border-white"><Play fill="white" size={24} /></div>
                                                </div>
                                            </div>
                                            <div className="flex flex-col justify-center flex-1">
                                                <div className="flex items-start justify-between gap-3 mb-1">
                                                    <span className="text-lg font-bold cursor-pointer hover:underline" onClick={() => navigate(`/watch/${ep.id}`)}>{ep.title}</span>
                                                    {isAdmin && (
                                                        <button onClick={() => handleDeleteEpisode(ep.id)} className="flex-shrink-0 text-red-500 hover:text-red-400 transition mt-1" title="Delete episode">
                                                            <Trash2 size={16} />
                                                        </button>
                                                    )}
                                                </div>
                                                <div className="text-sm text-gray-400 mb-2 font-medium">
                                                    Season {ep.season_number} • Episode {ep.episode_number}
                                                </div>
                                                <p className="text-gray-300 text-sm line-clamp-3">{ep.description}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <div className="text-zinc-500 py-10 text-center bg-zinc-900/30 rounded-lg">
                                    <Info className="w-12 h-12 mx-auto mb-4 opacity-30" />
                                    No episodes available yet.
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {/* Similar Movies Section */}
            {similarMovies.length > 0 && (
                <div className="px-4 md:px-12 py-8 max-w-7xl mx-auto">
                    <h2 className="text-2xl font-bold mb-6 border-b border-zinc-800 pb-2">Similar Titles</h2>
                    <div className="flex gap-4 overflow-x-auto pb-4 scroll-smooth" style={{ scrollBehavior: 'smooth' }}>
                        {similarMovies.map(movie => (
                            <div
                                key={movie.id}
                                onClick={() => navigate(`/title/${movie.id}`)}
                                className="group flex-shrink-0 w-48 cursor-pointer"
                            >
                                <div className="relative w-full aspect-video bg-black rounded-lg overflow-hidden mb-3 border border-zinc-800 group-hover:border-netflix-red transition">
                                    <img
                                        src={movie.thumbnail_url || `https://placehold.co/320x180/111/333?text=${encodeURIComponent(movie.title)}`}
                                        alt={movie.title}
                                        className="w-full h-full object-cover group-hover:brightness-75 transition duration-300"
                                    />
                                    <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                        <div className="bg-netflix-red rounded-full p-3 border-2 border-white">
                                            <Play fill="white" size={24} />
                                        </div>
                                    </div>
                                </div>
                                <h3 className="text-sm font-semibold text-white group-hover:text-netflix-red transition line-clamp-2">
                                    {movie.title}
                                </h3>
                                {movie.release_year && (
                                    <p className="text-xs text-gray-400 mt-1">{movie.release_year}</p>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

export default TitleDetailsPage;
