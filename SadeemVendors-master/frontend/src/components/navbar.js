import React, { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {jwtDecode} from 'jwt-decode'; // Correct import for jwt-decode
import '../css/Navbar.css';
import logo from '../css/vendor.jpg';
import defaultImage from '../css/vendor.jpg';
import { useOrderUpdate } from './OrderUpdateContext'; // Import the custom hook

const Navbar = ({ initialCartItems, onCartItemsChange, refreshCart }) => {
    const [userRole, setUserRole] = useState(null);
    const [isTransparent, setIsTransparent] = useState(false);
    const [cartDropdownVisible, setCartDropdownVisible] = useState(false);
    const [cartItems, setCartItems] = useState(initialCartItems || []);
    const [errorMessage, setErrorMessage] = useState(null);
    const [removalErrorMessage, setRemovalErrorMessage] = useState(null);
    const [totalPrice, setTotalPrice] = useState(0);
    const [totalQuantity, setTotalQuantity] = useState(0);
    const [loadingCartItems, setLoadingCartItems] = useState(false);
    const [loadingCheckout, setLoadingCheckout] = useState(false);
    const navigate = useNavigate();
    const dropdownRef = useRef(null);
    const { setShouldUpdateOrders } = useOrderUpdate(); // Use the context

    useEffect(() => {
        fetchCartItems();
    }, [refreshCart]);

    useEffect(() => {
        setLoadingCartItems(true);

        const token = localStorage.getItem('token');
        if (token) {
            try {
                const decodedToken = jwtDecode(token);
                const currentTime = Date.now() / 1000;
                if (decodedToken.exp < currentTime) {
                    localStorage.removeItem('token');
                    setUserRole(null);
                } else {
                    setUserRole(decodedToken.userRole);
                    fetchCartItems();
                }
            } catch (error) {
                console.error('Error decoding token:', error);
                localStorage.removeItem('token');
                setUserRole(null);
            }
        }
    }, []);

    useEffect(() => {
        const handleScroll = () => {
            setIsTransparent(window.scrollY > 50);
        };

        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setCartDropdownVisible(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const fetchCartItems = async () => {
        const token = localStorage.getItem('token');
        if (!token) return;

        try {
            const [itemsResponse, cartResponse] = await Promise.all([
                fetch('http://localhost:8080/cartitems', {
                    headers: { 'Authorization': `Bearer ${token}` }
                }),
                fetch('http://localhost:8080/carts', {
                    headers: { 'Authorization': `Bearer ${token}` }
                })
            ]);

            if (!itemsResponse.ok || !cartResponse.ok) {
                throw new Error('You must have a table to checkout!');
            }

            const itemsData = await itemsResponse.json();
            setCartItems(itemsData.cart || []);
            console.log('Cart items:', itemsData.cart);

            const cartData = await cartResponse.json();
            console.log('Cart data:', cartData);

            setTotalPrice(cartData.cart?.total_price || 0);
            setTotalQuantity(cartData.cart?.quantity || 0);
            console.log('Fetched cart items:', itemsData.cart);

        } catch (error) {
            console.error('Error fetching cart items:', error);
        } finally {
            setLoadingCartItems(false);
        }
    };

    const handleSignOut = () => {
        localStorage.removeItem('token');
        setUserRole(null);
        navigate('/signin');
    };

    const toggleCartDropdown = () => {
        setCartDropdownVisible(prev => !prev);
        setErrorMessage(null);
        setRemovalErrorMessage(null);
    };
    const updateCartItemQuantity = async (itemId, newQuantity) => {
        if (newQuantity < 1) return;
    
        setErrorMessage(null);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/cartitems/${itemId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    'Authorization': `Bearer ${token}`
                },
                body: new URLSearchParams({ item_id: itemId, quantity: newQuantity }).toString()
            });
    
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
    
            await fetchCartItems();
        } catch (error) {
            setErrorMessage('Error updating quantity. Please try again.');
            console.error('Error updating cart item:', error);
        } 
    };

    const deleteCartItem = async (itemId) => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`http://localhost:8080/cartitems/${itemId}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    'Authorization': `Bearer ${token}`
                }
            });

            if (!response.ok) {
                throw new Error('Network response was not ok');
            }

            console.log('Item deleted successfully');
            await fetchCartItems();
        } catch (error) {
            setRemovalErrorMessage('Error removing item. Please try again.');
            console.error('Error deleting cart item:', error);
        }
    };
    const handleCheckout = async () => {
        setLoadingCheckout(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch('http://localhost:8080/checkout', {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` }
            });
        
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.errors ? JSON.stringify(errorData.errors) : 'Network response was not ok');
            }
        
            console.log('Checkout successful');
            setCartItems([]);
            setShouldUpdateOrders(true); // Notify context that orders need to be updated
            window.location.reload(); // Refresh the page
        } catch (error) {
            console.error('Error during checkout:', error);
            // Optionally, you could also set an error message in the state here
            setErrorMessage(error.message); // Assuming you have a state for error messages
        } finally {
            setLoadingCheckout(false);
        }
    };
    
    return (
        <nav className={`navbar ${isTransparent ? 'transparent' : ''}`}>
            <div className="logo">
                <img src={logo} alt="Logo" />
            </div>
            <ul className="center-links">
                <li><Link to="/">Home</Link></li>
            </ul>
            <ul className="end-links">
                {userRole === "1" && (
                    <>
                        <li><Link to="/add-vendor">Add Vendor</Link></li>
                        <li><Link to="/users">Users</Link></li>
                    </>
                )}
                {userRole ? (
                    <>
                        <li><Link to="/profile">Profile</Link></li>
                        <li><Link to="/orders">Orders</Link></li>
                        <li>
                            <button onClick={toggleCartDropdown} disabled={loadingCartItems}>
                                {loadingCartItems ? 'Loading Cart...' : 'Cart'}
                            </button>
                            {cartDropdownVisible && (
                                <div className="cart-dropdown" ref={dropdownRef}>
                                    {loadingCartItems ? (
                                        <p>Loading cart items...</p>
                                    ) : cartItems.length === 0 ? (
                                        <p>No items in cart</p>
                                    ) : (
                                        <>
                                            <ul className="cart-list">
                                                {cartItems.map(item => (
 <li key={item.item_id} className="cart-item">
                                                            <div className="cart-item-img">
                                                                {console.log(item.img)}
                                                            <img src={item.img || defaultImage} alt={item.name} />                                                        </div>
                                                        <div className="cart-item-name">{item.name}</div>
                                                        <div className="cart-item-quantity">
                                                            <button onClick={() => updateCartItemQuantity(item.item_id, item.quantity - 1)}>-</button>
                                                            <span>{item.quantity}</span>
                                                            <button onClick={() => updateCartItemQuantity(item.item_id, item.quantity + 1)}>+</button>
                                                        </div>
                                                        <button onClick={() => deleteCartItem(item.item_id)} className='removebutton'>Remove</button>
                                                    </li>
                                                ))}
                                            </ul>
                                            <div className="cart-summary">
                                                <p>Total Quantity: {totalQuantity}</p>
                                                <p>Total Price: ${totalPrice.toFixed(2)}</p>
                                            </div>
                                        </>
                                    )}
                                    {errorMessage && <p className="error-message">{errorMessage}</p>}
                                    {removalErrorMessage && <p className="error-message">{removalErrorMessage}</p>}
                                    <button onClick={handleCheckout} disabled={loadingCheckout}>
                                        {loadingCheckout ? 'Processing Checkout...' : 'Checkout'}
                                    </button>
                                </div>
                            )}
                        </li>
                        <li><button onClick={handleSignOut}>Sign Out</button></li>
                    </>
                ) : (
                    <li><Link to="/signin">Sign In</Link></li>
                )}
            </ul>
        </nav>
    );
};

export default Navbar;
