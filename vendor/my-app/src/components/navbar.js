import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { jwtDecode } from 'jwt-decode'; // Correct import
import '../css/Navbar.css'; // For styling

const Navbar = () => {
    const navigate = useNavigate();
    const token = localStorage.getItem('token');
    const userRole = token ? jwtDecode(token).userRole : null;

    const handleSignOut = () => {
        localStorage.removeItem('token');
        navigate('/signin');
    };

    return (
        <nav className="navbar">
            <ul className="center-links">
                <li><Link to="/">Home</Link></li>
            </ul>
            <ul className="end-links">
                {userRole === "1" && (
                    <li><Link to="/add-vendor">Add Vendor</Link></li>
                )}
                {userRole && (
                    <>
                        <li><Link to="/profile">Profile</Link></li>
                        <li><button onClick={handleSignOut}>Sign Out</button></li>
                    </>
                )}
                {!userRole && (
                    <li><Link to="/signin">Sign In</Link></li>
                )}
            </ul>
        </nav>
    );
};

export default Navbar;
