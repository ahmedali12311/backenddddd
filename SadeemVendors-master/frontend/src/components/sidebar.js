// src/components/Sidebar.js
import React, { useState, useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import '../css/Sidebar.css'; // For styling
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faBars, faTimes } from '@fortawesome/free-solid-svg-icons'; // Import icons

const Sidebar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [userRole, setUserRole] = useState(null); // State to store user role
  const sidebarRef = useRef(null);

  const toggleSidebar = () => {
    setIsOpen(prev => !prev);
  };

  const handleClickOutside = (event) => {
    if (sidebarRef.current && !sidebarRef.current.contains(event.target)) {
      setIsOpen(false);
    }
  };

  useEffect(() => {
    // Add event listener to detect clicks outside
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      // Clean up event listener on component unmount
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  useEffect(() => {
    const fetchUserRole = async () => {
      try {
        const token = localStorage.getItem('token');
        const response = await fetch('http://localhost:8080/me', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          if (response.status === 401) {
            // Handle unauthorized access
            return;
          }
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        setUserRole(data.me.user_role);
      } catch (error) {
        console.error('Error fetching user role:', error);
      }
    };

    fetchUserRole();
  }, []);

  return (
    <div>
      <div ref={sidebarRef} className={`sidebar ${isOpen ? 'open' : ''}`}>
        <div className="sidebar-content">
          <ul>
            <li><Link to="/vendors">Vendors</Link></li>
            {userRole === '1' && (
              <li><Link to="/users">Users</Link></li>
            )}
          </ul>
        </div>
        <button className="sidebar-toggle" onClick={toggleSidebar}>
          <FontAwesomeIcon 
            icon={isOpen ? faTimes : faBars} 
          />
        </button>
      </div>
      {isOpen && (
        <div className="overlay" onClick={toggleSidebar}></div>
      )}
    </div>
  );
};

export default Sidebar;
