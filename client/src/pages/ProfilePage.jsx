import { useAuth } from '../context/AuthContext';

const ProfilePage = () => {
    const { user, logout } = useAuth();

    return (
        <div className="min-h-screen bg-netflix-dark pt-24 px-4 md:px-12 flex justify-center">
            <div className="max-w-2xl w-full">
                <h1 className="text-4xl font-bold mb-8 border-b border-gray-700 pb-4">Account</h1>

                <div className="bg-zinc-900 rounded p-6 flex gap-6 items-start">
                    <div className="w-24 h-24 bg-blue-600 rounded flex items-center justify-center text-4xl font-bold">
                        {user?.name?.[0] || 'U'}
                    </div>

                    <div className="flex-grow">
                        <div className="flex items-center gap-3 mb-2">
                            <h2 className="text-2xl font-semibold">{user?.name}</h2>
                            {user?.role === 'admin' && (
                                <span className="bg-netflix-red text-white text-[10px] uppercase font-bold px-2 py-0.5 rounded">Admin</span>
                            )}
                        </div>
                        <p className="text-gray-400 mb-4">Email: {user?.email}</p>
                        <p className="text-gray-400 mb-4">Member since using Go StreamApp</p>

                        <button
                            onClick={logout}
                            className="bg-gray-800 border border-gray-600 px-4 py-2 text-white font-semibold hover:bg-gray-700 transition"
                        >
                            Sign Out
                        </button>
                    </div>
                </div>

                <div className="mt-8">
                    <h3 className="text-xl font-bold mb-4 text-gray-300">Plan Details</h3>
                    <div className="bg-zinc-900 rounded p-4 flex justify-between items-center">
                        <div>
                            <span className="font-bold text-netflix-red">Premium</span> <span className="text-xl font-bold">4K + HDR</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ProfilePage;
