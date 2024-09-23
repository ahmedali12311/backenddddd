// src/pages/UsersPage.js
import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import '../css/users.css';
import defaultImageProfile from '../css/profile.jpg';

const UsersPage = () => {
  const [users, setUsers] = useState([]);
  const [dropdownResults, setDropdownResults] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortOrder, setSortOrder] = useState('latest');
  const [page, setPage] = useState(1);
  const pageSize = 12;
  const navigate = useNavigate(); // Initialize useNavigate

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('token');
        const response = await fetch(`http://localhost:8080/users?page=${page}&pageSize=${pageSize}&sortColumn=${sortOrder.split('_')[0]}&sortDirection=${sortOrder.split('_')[1]}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (!response.ok) {
          throw new Error('Failed to fetch users');
        }
        const data = await response.json();
        console.log('Fetched users data:', data); // Log the entire fetched data
        setUsers(data.users || []);
      } catch (error) {
        setError(error.message);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, [sortOrder, page]);

  useEffect(() => {
    if (searchTerm) {
      const results = users.filter(user =>
        user.name.toLowerCase().includes(searchTerm.toLowerCase())
      );
      console.log('Dropdown results:', results); // Log search results
      setDropdownResults(results);
    } else {
      setDropdownResults([]);
    }
  }, [searchTerm, users]);

  const handleSearchChange = (event) => {
    setSearchTerm(event.target.value);
    setPage(1);
  };

  const handleSortChange = (event) => {
    setSortOrder(event.target.value);
    setPage(1);
  };

  const handlePageChange = (newPage) => {
    setPage(newPage);
  };

  const handleImageError = (e) => {
    e.target.src = defaultImageProfile; // Use default image on error
  };

  const handleSearchResultClick = (userId) => {
    // Navigate to the edit page for the selected user
    navigate(`/users/edit/${userId}`);
  };


  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;


  return (
    <div className="user-list-container">
      <h1 className="title">Users</h1>
      <div className="search-input-container">
        <input
          type="text"
          className="search-input"
          placeholder="Search users..."
          value={searchTerm}
          onChange={handleSearchChange}
        />
    {searchTerm && (
  <ul className="dropdown-menu">
    {dropdownResults.length > 0 ? (
      dropdownResults.map(user => (
        <li 
          key={user.id} 
          onClick={() => handleSearchResultClick(user.id)} // Navigate on click
          className="dropdown-item"
        >
          <img 
            src={user.img && user.img.trim() ? user.img : defaultImageProfile} 
            alt={user.name} 
            className="dropdown-user-image"
            onError={handleImageError} // Fallback to default image if there's an error
          />
          <span>{user.name}</span>
        </li>
      ))
    ) : (
      <li>No results found</li>
    )}
  </ul>
)}
      </div>
      <div className="sort-selection">
        <select value={sortOrder} onChange={handleSortChange}>
          <option value="latest_ASC">Latest (Ascending)</option>
          <option value="latest_DESC">Latest (Descending)</option>
          <option value="name_asc">Name (Ascending)</option>
          <option value="name_desc">Name (Descending)</option>
        </select>
      </div>
      <div className="user-list">
        {users.length > 0 ? (
          users.map(user => (
            <div className="user-card" key={user.id}>
              <div className="user-header">
                <div className="user-image-container">
                  <img 
                    src={user.img && user.img.trim() ? user.img : defaultImageProfile} 
                    alt={user.name} 
                    className="user-image"
                    onError={handleImageError} // Fallback to default image if there's an error
                  />
                </div>
                <div className="user-content">
                  <h2 className="user-name">{user.name}</h2>
                  <p className="user-phone">Phone: {user.phone && user.phone.trim() ? user.phone : 'N/A'}</p>
                  <p className="user-email">Email: {user.email && user.email.trim() ? user.email : 'N/A'}</p>
                  <Link to={`/users/edit/${user.id}`} className="edit-link">Edit Profile</Link>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div>No users found</div>
        )}
      </div>
      <div className="pagination">
        <button 
          onClick={() => handlePageChange(page - 1)} 
          disabled={page === 1}
        >
          Previous
        </button>
        <span>Page {page}</span>
        <button 
          onClick={() => handlePageChange(page + 1)} 
          disabled={users.length < pageSize}
        >
          Next
        </button>
      </div>
    </div>
  );
};

export default UsersPage;
