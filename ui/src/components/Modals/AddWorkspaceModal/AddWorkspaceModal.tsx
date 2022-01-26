import React, { FC, useEffect, useState } from 'react';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
// import './AddUserModal.sass';

interface AddWorkspaceModal {
  isOpen : boolean
}

const AddWorkspaceModal: FC<AddWorkspaceModal> = ({isOpen,children}) => {

  const [isOpenModal,setIsOpen] = useState(isOpen)

  useEffect(() => {
    setIsOpen(isOpen)
  },[isOpen])

  const onClose = () => {}

  const onConfirm = () => {}

  return (<>
    <ConfirmationModal isOpen={isOpenModal} onClose={onClose} onConfirm={onConfirm} title=''>
      {children}
    </ConfirmationModal>
    </>); 
};

export default AddWorkspaceModal;
