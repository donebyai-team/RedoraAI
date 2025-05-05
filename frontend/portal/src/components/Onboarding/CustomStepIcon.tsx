import React from 'react';
import { styled } from '@mui/material/styles';
import { Mail, ShieldCheck, SendHorizonal, FileText } from 'lucide-react';
import { StepIconProps } from '@mui/material';

// Create a styled component for the icon container
const IconContainer = styled('div')<{
  ownerState: { active?: boolean; completed?: boolean; error?: boolean };
}>(({ theme, ownerState }) => ({
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  width: 40,
  height: 40,
  borderRadius: '50%',
  backgroundColor:
    ownerState.completed ? theme.palette.primary.main :
      ownerState.active ? theme.palette.primary.light :
        theme.palette.grey[100],
  color:
    ownerState.completed ? theme.palette.primary.contrastText :
      ownerState.active ? theme.palette.primary.contrastText :
        theme.palette.text.secondary,
  transition: theme.transitions.create(['background-color', 'color', 'box-shadow'], {
    duration: theme.transitions.duration.shorter,
  }),
  ...(ownerState.active && {
    boxShadow: `0 4px 10px 0 ${theme.palette.primary.light}80`,
  }),
}));

// interface CustomStepIconProps extends StepIconProps {
//   active?: boolean;
//   completed?: boolean;
//   error?: boolean;
//   icon: number;
// }

const CustomStepIcon: React.ElementType<StepIconProps> = (props) => {
  const { active, completed, icon } = props;
  const ownerState = { active, completed };

  const icons: { [index: number]: React.ReactElement } = {
    1: <Mail size={20} />,
    2: <ShieldCheck size={20} />,
    3: <SendHorizonal size={20} />,
    4: <FileText size={20} />,
  };

  return (
    <IconContainer ownerState={ownerState}>
      {completed ? <ShieldCheck size={20} /> : icons[icon as number]}
    </IconContainer>
  );
};

export default CustomStepIcon;