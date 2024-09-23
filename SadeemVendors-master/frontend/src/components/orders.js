import React, { useEffect, useState, useCallback } from 'react';
import '../css/orders.css'; // Import the updated CSS

const Orders = () => {
    const [orders, setOrders] = useState([]);
    const [error, setError] = useState(null);
    const [userRole, setUserRole] = useState(null);
    const fetchOrders = useCallback(async () => {
        try {
            const token = localStorage.getItem('token');
            if (!token) {
                throw new Error('No token found in localStorage');
            }
    
            const decodedToken = JSON.parse(atob(token.split('.')[1]));
            const role = decodedToken.userRole;
            const userId = decodedToken.id;
    
            console.log('Decoded Token:', decodedToken); // Log token data
            console.log('Role:', role);
            console.log('User ID:', userId);
    
            setUserRole(role);
    
            let url = '/orders';
    
            if (role === '2') {
                // Role '2' indicates vendor owner: fetch vendor ID first
                const vendorResponse = await fetch(`http://localhost:8080/uservendors/${userId}`, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
    
                console.log('Vendor Response Status:', vendorResponse.status); // Log vendor response status
    
                if (!vendorResponse.ok) {
                    throw new Error(`Failed to fetch vendor ID: ${vendorResponse.statusText}`);
                }
    
                const vendorData = await vendorResponse.json();
                console.log('Vendor Data:', vendorData); // Log vendor data
    
                const vendor = vendorData.vendor[0]; // Access the first vendor in the array
                const vendorId = vendor?.id; // Safely access the id
    
                if (!vendorId) {
                    throw new Error('Vendor ID is undefined');
                }
    
                // Include the vendor ID in the request to /orders endpoint
                url = `/vendororders/${vendorId}`;
            }
    
            console.log(`Fetching orders from: ${url}`); // Log final URL
    
            const response = await fetch(`http://localhost:8080${url}`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
    
            console.log(`Orders Response Status: ${response.status}`); // Log orders response status
    
            if (!response.ok) {
                throw new Error(`Network response was not ok: ${response.statusText}`);
            }
    
            const data = await response.json();
            console.log('Fetched orders:', data); // Log fetched orders
            setOrders(data.orders || []);
        } catch (error) {
            setError('Error fetching orders. Please try again.');
            console.error('Error fetching orders:', error);
        }
    }, []);
    
    
    useEffect(() => {
        fetchOrders(); // Call fetchOrders on component mount
    }, [fetchOrders]); // Adding fetchOrders to dependency array
    const handleCompleteOrder = async (orderId) => {
        try {
            const response = await fetch(`http://localhost:8080/orderscompleted/${orderId}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/x-www-form-urlencoded' // Set Content-Type for form-urlencoded
                },
                body: new URLSearchParams({
                    status: 'completed' // Format body as URLSearchParams
                })
            });
    
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
    
            // Refresh the orders after updating
            await fetchOrders();
        } catch (error) {
            setError('Error updating order status. Please try again.');
            console.error('Error updating order status:', error);
        }
    };
    
    return (
        <div className="orders-container">
            <div className="orders-card">
                <div className="orders-header">Orders</div>
                {error && <p className="error-message">{error}</p>}
                <div className="orders-list">
                    {orders.map(order => (
                        <div key={order.id} className="order-item">
                            <div className="order-details">
                                <strong>Total Cost:</strong> ${order.total_order_cost}
                            </div>
                            <div className="order-details">
                                <strong>Status:</strong> {order.status}
                            </div>
                            <div className="order-vendor">
                                <strong>Vendor Name:</strong> {order.vendor_name}
                            </div>
                            <div className="order-items">
                                <strong>Items:</strong>
                                {order.item_names && order.item_names.map((item, index) => (
                                    <div key={index} className="order-item-details">
                                        <span>{item}</span>
                                        <span>${order.item_prices[index]}</span>
                                    </div>
                                ))}
                            </div>
                            {userRole === '2' && order.status !== 'completed' && (
                                <button onClick={() => handleCompleteOrder(order.id)}>Mark as Completed</button>
                            )}
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default Orders;
