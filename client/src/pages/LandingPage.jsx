import { Link } from 'react-router-dom';

const LandingPage = () => {
    return (
        <div className="relative h-screen w-full bg-[url('https://assets.nflxext.com/ffe/siteui/vlv3/c38a2d52-138e-48a3-ab68-36787ece46b3/eeb03fc9-99bf-47d3-a6a9-a0ca713e3146/US-en-20240101-popsignuptwoweeks-perspective_alpha_website_large.jpg')] bg-cover bg-center">
            <div className="absolute inset-0 bg-black/60 bg-gradient-to-t from-netflix-black via-transparent to-black/50"></div>

            <div className="relative z-10 flex flex-col items-center justify-center h-full text-center px-4">
                <h1 className="text-5xl md:text-6xl font-bold mb-4 max-w-4xl">Unlimited movies, TV shows, and more</h1>
                <p className="text-xl md:text-2xl mb-8">Watch anywhere. Cancel anytime.</p>
                <p className="text-lg md:text-xl mb-6">Ready to watch? Enter your email to create or restart your membership.</p>

                <Link to="/login" className="bg-netflix-red text-white text-xl px-8 py-3 rounded hover:bg-red-700 transition font-medium flex items-center gap-2">
                    Get Started
                </Link>
            </div>

            <div className="absolute top-0 w-full p-6 flex justify-between items-center z-20">
                <div className="text-netflix-red text-4xl font-bold">GO STREAM</div>
                <Link to="/login" className="bg-netflix-red text-white px-4 py-1.5 rounded text-sm font-semibold hover:bg-red-700">
                    Sign In
                </Link>
            </div>
        </div>
    );
};

export default LandingPage;
