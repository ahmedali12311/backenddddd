import React, { useState } from 'react';
import { Routes, Route, BrowserRouter as Router, useLocation } from 'react-router-dom';
import Login from './components/login.js';
import Vendors from './components/vendors.js';
import VendorDetails from './components/vendordetails.js';
import AddVendor from './components/addvendor.js';
import Navbar from './components/navbar.js';
import Sidebar from './components/sidebar.js';
import Profile from './components/editprofile.js';
import UsersPage from './components/userspage.js';
import EditUser from './components/edituser.js';
import Editvendorer from './components/editvendor.js';
import Orders from './components/orders.js';
import { OrderUpdateProvider } from './components/OrderUpdateContext'; // Import the new context provider

function App() {
  const [cartItems, setCartItems] = useState([]);
  const location = useLocation();

  // List of routes where Navbar and Sidebar should not be shown
  const noNavAndSidebarRoutes = ['/signin', '/signup'];
  const [refreshCart, setRefreshCart] = useState(false); // New state for triggering refresh

  const handleAddToCart = (item) => {
    setCartItems((prevCartItems) => [...prevCartItems, item]);
    setRefreshCart((prev) => !prev); // Toggle the state to trigger a refresh
};


  return (
    <div>
      {!noNavAndSidebarRoutes.includes(location.pathname) && (
        <>
            <Navbar cartItems={cartItems} onCartItemsChange={handleAddToCart} refreshCart={refreshCart} />
            <Sidebar />
        </>
      )}

      <div className="main-content">
        <Routes>
          <Route path="/signin" element={<Login />} />
          <Route path="/" element={<Vendors />} />
          <Route path="/vendors" element={<Vendors />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/orders" element={<Orders />} />

          <Route path="/vendor/:id" element={<VendorDetails onAddToCart={handleAddToCart}/>} />
          <Route path="/edit-vendor/:id" element={<Editvendorer />} />

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
        <OrderUpdateProvider>
        <App />
        </OrderUpdateProvider>
    </Router>
  );
}