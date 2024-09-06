import React, { useState, useEffect, useRef } from 'react';
import '../css/editprofile.css';
import defaultImage from './vendor.jpg';

function EditProfile() {
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState(null);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [phone, setPhone] = useState('');
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
  const imageRef = useRef(null);
  const userId = localStorage.getItem('userId');

  useEffect(() => {
    const fetchUserDetails = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          throw new Error('No token found');
        }

        const response = await fetch('http://localhost:8080/me', {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();
        console.log('Fetched user data:', data);

        if (data.me && data.me.user_info) {
          const { name, email, phone, img } = data.me.user_info;
          setName(name || '');
          setEmail(email || '');
          setPhone(phone || '');
          setPreview(img || defaultImage);

          localStorage.setItem('userId', data.me.user_info.id);
        } else {
          console.error('No user_info in response:', data);
          setErrorMessages(prev => ({ ...prev, general: 'User data is missing in response' }));
        }
      } catch (error) {
        console.error('Error fetching user details:', error);
        setErrorMessages(prev => ({ ...prev, general: 'Failed to load user details' }));
      } finally {
        setLoading(false);
      }
    };

    fetchUserDetails();
  }, []);
  const handleImageClick = (event) => {
    const file = event.target.files[0];
    if (file) {
      // Check if the file is an image
      const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
      if (!validTypes.includes(file.type)) {
        setErrorMessages(prev => ({
          ...prev,
          image: 'Invalid image type. Please upload a JPEG, PNG, GIF, or WEBP image.'
        }));
        
        // Do not clear the current image or preview
        setImage(null);
        setPreview(prev => prev || defaultImage);
  
        // Hide the error message after 3 seconds
        setTimeout(() => {
          setErrorMessages(prev => ({ ...prev, image: '' }));
        }, 10000);
  
        return;
      }
  
      // Check if the image size is valid
      if (file.size > 2000000) { // 2MB
        setErrorMessages(prev => ({
          ...prev,
          image: 'Image size must be less than 2MB.'
        }));
  
        // Do not clear the current image or preview
        setImage(null);
        setPreview(prev => prev || defaultImage);
  
        // Hide the error message after 3 seconds
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
    formData.append('password', password);
    formData.append('phone', phone);
  
    if (image) {
      formData.append('img', image);
    }
  
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8080/users/${userId}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });
  
      if (!response.ok) {
        const errorData = await response.json();
        console.error('Backend errors:', errorData);
  
        // Update error messages based on backend response
        setErrorMessages(prev => ({
          ...prev,
          ...(errorData.error || {})
        }));
  
        // If the backend error is related to email
        if (response.status === 409 && errorData.error === 'Email already exists, try something else') {
          setErrorMessages(prev => ({ ...prev, email: 'Email already exists, try something else.' }));
        } else if (response.status === 422 && errorData.error) {
          // Handle specific validation errors
          setErrorMessages(prev => ({
            ...prev,
            ...errorData.error // Assuming `errorData.error` is an object like { phone: 'Invalid phone number' }
          }));
        } else {
          setErrorMessages(prev => ({ ...prev, general: 'Failed to update profile' }));
        }
  
        throw new Error('Failed to update profile');
      }
  
      setSuccess('Profile updated successfully');
      setTimeout(() => setSuccess(null), 4000); 
    } catch (error) {
      console.error('Error updating profile:', error);
    } finally {
      setLoading(false);
    }
  };
  
  

  if (loading) {
    return <div>Loading...</div>;
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
        <h1>Edit Profile</h1>
        <form onSubmit={handleSave} encType="multipart/form-data">
          <div className={`form-group ${errorMessages.name ? 'error' : ''}`}>
            <label htmlFor="name">Name</label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              style={{ borderColor: errorMessages.name ? 'red' : '' }}
            />
            {errorMessages.name && (
              <div className="error-message">{errorMessages.name}</div>
            )}
          </div>
         <div className={`form-group ${errorMessages.email ? 'error' : ''}`}>
  <label htmlFor="email">Email</label>
  <input
    type="email"
    id="email"
    value={email}
    onChange={(e) => setEmail(e.target.value)}
    style={{ borderColor: errorMessages.email ? 'red' : '' }}
  />
  {errorMessages.email && (
    <div className="error-message">{errorMessages.email}</div>
  )}
</div>
          <div className={`form-group ${errorMessages.phone ? 'error' : ''}`}>
            <label htmlFor="phone">Phone</label>
            <input
              type="text"
              id="phone"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              style={{ borderColor: errorMessages.phone ? 'red' : '' }}
            />
            {errorMessages.phone && (
              <div className="error-message">{errorMessages.phone}</div>
            )}
          </div>
          <div className={`form-group ${errorMessages.password ? 'error' : ''}`}>
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={{ borderColor: errorMessages.password ? 'red' : '' }}
            />
            {errorMessages.password && (
              <div className="error-message">{errorMessages.password}</div>
            )}
          </div>
          <button type="submit" disabled={loading}>Save Changes</button>
          {success && <div className="success-message">{success}</div>}
        </form>
      </div>
    </div>
  );
  
}

export default EditProfile;
