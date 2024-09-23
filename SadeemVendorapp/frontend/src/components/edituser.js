import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import '../css/editprofile.css';
import defaultImage from '../css/vendor.jpg';

function EditUser() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState(null);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [phone, setPhone] = useState('');
  const [vendorId, setVendorId] = useState('');
  const [role, setRole] = useState('');
  const [currentRole, setCurrentRole] = useState(''); // Store current role
  const [errorMessages, setErrorMessages] = useState({
    name: '',
    email: '',
    password: '',
    phone: '',
    image: '',
    general: ''
  });
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(null);
  const [successColor, setSuccessColor] = useState('');
  const imageRef = useRef(null);

  useEffect(() => {
    const fetchUserDetails = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          throw new Error('No token found');
        }

        const [userResponse, roleResponse] = await Promise.all([
          fetch(`http://localhost:8080/users/${userId}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }),
          fetch(`http://localhost:8080/userroles/${userId}`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })
        ]);

        if (!userResponse.ok || !roleResponse.ok) {
          const errorData = await userResponse.json() || {};
          console.error('Backend errors:', errorData);

          const newErrors = {
            name: errorData.error?.name || '',
            email: errorData.error?.email || '',
            password: errorData.error?.password || '',
            phone: errorData.error?.phone || '',
            image: '',
            general: '',
          };

          if (userResponse.status === 409 && errorData.error === 'Email already exists, try something else') {
            newErrors.email = 'Email already exists, try something else.';
          }

          setErrorMessages(newErrors);
          throw new Error('Failed to fetch user details or role information');
        }

        const userData = await userResponse.json();
        const roleData = await roleResponse.json();

        if (userData.user && roleData.user_roles) {
          const { name, email, phone, img } = userData.user;
          const { roleID } = roleData.user_roles;

          setName(name || '');
          setEmail(email || '');
          setPhone(phone || '');
          setPreview(img || defaultImage);
          setRole(roleID);
          setCurrentRole(roleID); // Set current role for comparison
        } else {
          console.error('No user or role data in response');
          setErrorMessages(prev => ({ ...prev, general: 'User or role data is missing in response' }));
        }
      } catch (error) {
        console.error('Error fetching user details:', error);
        setErrorMessages(prev => ({ ...prev, general: 'Failed to load user details' }));
      } finally {
        setLoading(false);
      }
    };

    fetchUserDetails();
  }, [userId]);

  const handleImageClick = (event) => {
    const file = event.target.files[0];
    if (file) {
      const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
      if (!validTypes.includes(file.type)) {
        setErrorMessages(prev => ({
          ...prev,
          image: 'Invalid image type. Please upload a JPEG, PNG, GIF, or WEBP image.'
        }));
        
        setImage(null);
        setPreview(prev => prev || defaultImage);

        setTimeout(() => {
          setErrorMessages(prev => ({ ...prev, image: '' }));
        }, 10000);

        return;
      }

      if (file.size > 2000000) { // 2MB
        setErrorMessages(prev => ({
          ...prev,
          image: 'Image size must be less than 2MB.'
        }));

        setImage(null);
        setPreview(prev => prev || defaultImage);

        setTimeout(() => {
          setErrorMessages(prev => ({ ...prev, image: '' }));
        }, 3000);

        return;
      }

      setErrorMessages(prev => ({ ...prev, image: '' }));
      setImage(file);

      const reader = new FileReader();
      reader.onloadend = () => {
        setPreview(reader.result);
      };
      reader.readAsDataURL(file);
    }
  };
  const handleSave = async (event) => {
    event.preventDefault();
    
    setErrorMessages({
        name: '',
        email: '',
        password: '',
        phone: '',
        image: '',
        general: '',
    });
    
    let hasErrors = false;
    
    if (!name.trim()) {
        setErrorMessages(prev => ({ ...prev, name: 'Name is required.' }));
        hasErrors = true;
    }
    
    if (!email.trim()) {
        setErrorMessages(prev => ({ ...prev, email: 'Email is required.' }));
        hasErrors = true;
    }
    
    if (!phone.trim()) {
        setErrorMessages(prev => ({ ...prev, phone: 'Phone number is required.' }));
        hasErrors = true;
    }
    
    if (errorMessages.image) {
        hasErrors = true;
    }
    
    if (hasErrors) {
        setSuccess(null);
        return;
    }
    
    const formData = new FormData();
    formData.append('name', name);
    formData.append('email', email);
    formData.append('phone', phone);
    
    if (password.trim()) {
        formData.append('password', password);
    }
    
    if (image) {
        formData.append('img', image);
    }

    setLoading(true);

    try {
        const token = localStorage.getItem('token');

        // First update user details
        const userResponse = await fetch(`http://localhost:8080/users/${userId}`, {
            method: 'PUT',
            headers: {
                Authorization: `Bearer ${token}`,
            },
            body: formData,
        });

        if (!userResponse.ok) {
            const errorResponse = await userResponse.json();
            setErrorMessages(prev => ({ ...prev, general: errorResponse.error || 'Failed to update user details' }));
            throw new Error(`${errorResponse.error}`);
        }

        // If role has changed, update the role as well
        if (role !== currentRole) {
            const formDataRole = new URLSearchParams();
            formDataRole.append('role', role);
            formDataRole.append('vendorID', vendorId);
    
            const roleResponse = await fetch(`http://localhost:8080/grantrole/${userId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                    Authorization: `Bearer ${token}`,
                },
                body: formDataRole.toString(),
            });
    
            if (!roleResponse.ok) {
                const errorResponse = await roleResponse.json();
                setErrorMessages(prev => ({ ...prev, general: errorResponse.error || 'Failed to update user role' }));
                throw new Error(`${errorResponse.error}`);
            }
        }

        setSuccess('Profile updated successfully');
        setSuccessColor('green');
        setTimeout(() => setSuccess(null), 2000);
        setTimeout(() => navigate('/users'), 2000);
    } catch (error) {
        console.error('Error updating profile or role:', error);
        setErrorMessages(prev => ({ ...prev, general: error.message }));
    } finally {
        setLoading(false);
    }
};
  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete this user?')) {
      try {
        const token = localStorage.getItem('token');
        const response = await fetch(`http://localhost:8080/users/${userId}`, {
          method: 'DELETE',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          const errorData = await response.json();
          console.error('Backend errors:', errorData);
          setErrorMessages(prev => ({ ...prev, general: 'Failed to delete user' }));
          return;
        }

        setSuccess('User deleted successfully');
        setSuccessColor('red');
        setTimeout(() => {
          setSuccess(null);
          navigate('/users');
        }, 2000);
      } catch (error) {
        console.error('Error deleting user:', error);
        setErrorMessages(prev => ({ ...prev, general: 'Failed to delete user' }));
      }
    }
  };

  if (loading) {
    return <div>{loading && <div className="spinner"></div>}</div>;
  }

  const handleImageError = (e) => {
    e.target.src = defaultImage;
  };

  return (
    <div className="profile-container">
      <div className="profile-image-container">
        <input
          type="file"
          ref={imageRef}
          onChange={handleImageClick}
          style={{ display: 'none' }}
        />
        <img
          src={preview}
          alt={name}
          className={`profile-image ${!preview ? 'no-image' : ''}`}
          onError={handleImageError}
          onClick={() => imageRef.current?.click()}
        />
        {errorMessages.image && (
          <div className="error-message">{errorMessages.image}</div>
        )}
      </div>
      <div className="profile-info-container">
        <form onSubmit={handleSave}>
          <div className={`form-group ${errorMessages.name ? 'error' : ''}`}>
            <label htmlFor="name">Name:</label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
            {errorMessages.name && (
              <div className="error-message">{errorMessages.name}</div>
            )}
          </div>
          <div className={`form-group ${errorMessages.email ? 'error' : ''}`}>
            <label htmlFor="email">Email:</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            {errorMessages.email && (
              <div className="error-message">{errorMessages.email}</div>
            )}
          </div>
          <div className={`form-group ${errorMessages.password ? 'error' : ''}`}>
            <label htmlFor="password">Password:</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password" // Prevents browsers from autofilling this field

            />
            {errorMessages.password && (
              <div className="error-message">{errorMessages.password}</div>
            )}
          </div>
          <div className={`form-group ${errorMessages.phone ? 'error' : ''}`}>
            <label htmlFor="phone">Phone:</label>
            <input
              type="text"
              id="phone"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
            />
            {errorMessages.phone && (
              <div className="error-message">{errorMessages.phone}</div>
            )}
          </div>

          <div className={`form-group ${errorMessages.role ? 'error' : ''}`}>
            <label htmlFor="role">Role:</label>
            <input
              type="text"
              id="role"
              value={role}
              onChange={(e) => setRole(e.target.value)}
            />
            {errorMessages.role && (
              <div className="error-message">{errorMessages.role}</div>
            )}
          </div>
          {role === '2' && (
            <div className={`form-group ${errorMessages.vendorId ? 'error' : ''}`}>
              <label htmlFor="vendorId">Vendor ID:</label>
              <input
                type="text"
                id="vendorId"
                value={vendorId}
                onChange={(e) => setVendorId(e.target.value)}
              />
              {errorMessages.vendorId && (
                <div className="error-message">{errorMessages.vendorId}</div>
              )}
            </div>
          )}
          <button type="submit" className="save-button">Save Changes</button>
        </form>
        {errorMessages.general && (
          <div className="error-message">{errorMessages.general}</div>
        )}
        {success && (
          <div className="success-message" style={{ color: successColor }}>{success}</div>
        )}
        <button onClick={handleDelete} className="delete-button">Delete User</button>
      </div>
    </div>
  );
}

export default EditUser;