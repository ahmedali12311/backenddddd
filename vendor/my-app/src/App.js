import React from 'react';
import { Routes, Route, BrowserRouter as Router, useLocation } from 'react-router-dom';
import Login from './components/login.js';
import Vendors from './components/vendors.js';
import VendorDetails from './components/vendordetails.js';
import EditVendor from './components/editvendor.js';
import AddVendor from './components/addvendor.js';
import Navbar from './components/navbar.js';
import Sidebar from './components/sidebar.js';
import Profile from './components/editprofile.js';
import UsersPage from './components/userspage.js';
import EditUser from './components/edituser.js';

function App() {
  const location = useLocation();

  // List of routes where Navbar and Sidebar should not be shown
  const noNavAndSidebarRoutes = ['/signin', '/signup'];

  return (
    <div>
      {!noNavAndSidebarRoutes.includes(location.pathname) && (
        <>
          <Navbar />
          <Sidebar />
        </>
      )}

      <div className="main-content">
        <Routes>
          <Route path="/signin" element={<Login />} />
          <Route path="/" element={<Vendors />} />
          <Route path="/vendors" element={<Vendors />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/vendor/:id" element={<VendorDetails />} />
          <Route path="/edit-vendor/:id" element={<EditVendor />} />
          <Route path="/add-vendor" element={<AddVendor />} />
          <Route path="/users" element={<UsersPage />} />
          <Route path="/users/edit/:userId" element={<EditUser />} />


        </Routes>
      </div>
    </div>
  );
}

export default function AppWrapper() {
  return (
    <Router>
      <App />
    </Router>
  );
}