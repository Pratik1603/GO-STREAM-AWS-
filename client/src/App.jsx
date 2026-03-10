import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Navbar from './components/Navbar';
import LandingPage from './pages/LandingPage';
import AuthPage from './pages/AuthPage';
import BrowsePage from './pages/BrowsePage';
import PlayerPage from './pages/PlayerPage';
import ProfilePage from './pages/ProfilePage';
import UploadPage from './pages/UploadPage';
import TitleDetailsPage from './pages/TitleDetailsPage';

const ProtectedRoute = ({ children }) => {
  const { user, loading } = useAuth();
  if (loading) return <div>Loading...</div>;
  if (!user) return <Navigate to="/login" />;
  return children;
};

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="min-h-screen bg-netflix-dark text-white font-sans">
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/login" element={<AuthPage />} />
            <Route path="/register" element={<AuthPage isRegister />} />

            <Route path="/browse" element={
              <ProtectedRoute>
                <Navbar />
                <BrowsePage />
              </ProtectedRoute>
            } />

            <Route path="/title/:id" element={
              <ProtectedRoute>
                <Navbar />
                <TitleDetailsPage />
              </ProtectedRoute>
            } />

            <Route path="/watch/:id" element={
              <ProtectedRoute>
                <PlayerPage />
              </ProtectedRoute>
            } />

            <Route path="/profile" element={
              <ProtectedRoute>
                <Navbar />
                <ProfilePage />
              </ProtectedRoute>
            } />

            <Route path="/upload" element={
              <ProtectedRoute>
                <Navbar />
                <UploadPage />
              </ProtectedRoute>
            } />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
