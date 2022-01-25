import React, { lazy, Suspense } from 'react';

const LazyAddUserModal = lazy(() => import('./AddUserModal'));

const AddUserModal = (props: JSX.IntrinsicAttributes & { children?: React.ReactNode;isOpen:boolean }) => (
  <Suspense fallback={null}>
    <LazyAddUserModal {...props} />
  </Suspense>
);

export default AddUserModal;
