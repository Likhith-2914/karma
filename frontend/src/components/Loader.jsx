import React from 'react';

export const Spinner = ({ size = 24, color = 'var(--accent-primary)', className = '' }) => {
  return (
    <div 
      className={`spinner ${className}`} 
      style={{ 
        width: size, 
        height: size, 
        borderColor: `${color}40`, /* 40 is hex for 25% opacity */
        borderTopColor: color 
      }}
    ></div>
  );
};

export const FullPageLoader = ({ text = "Loading..." }) => {
  return (
    <div className="full-page-loader">
      <Spinner size={48} />
      {text && <p className="loader-text">{text}</p>}
    </div>
  );
};
