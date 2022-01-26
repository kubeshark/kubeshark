import React, { FC, useEffect, useState } from 'react';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import './AddUserModal.sass';

interface AddUserModalProps {
  isOpen : boolean
}

const AddUserModal: FC<AddUserModalProps> = ({isOpen}) => {

  const [isOpenModal,setIsOpen] = useState(isOpen)

  useEffect(() => {
    setIsOpen(isOpen)
  },[isOpen])

  const onClose = () => {}

  const onConfirm = () => {}

  return (<>
    <ConfirmationModal isOpen={isOpenModal} onClose={onClose} onConfirm={onConfirm} title=''>
      
    </ConfirmationModal>
    </>); 
};

export default AddUserModal;
