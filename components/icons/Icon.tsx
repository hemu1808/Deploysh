
import React from 'react';

interface IconProps {
    className?: string;
}

export const PlusIcon: React.FC<IconProps> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" className={className}><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
);

export const MinusIcon: React.FC<IconProps> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" className={className}><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 12H6" /></svg>
);

export const TrashIcon: React.FC<IconProps> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" className={className}><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
);

export const ExternalLinkIcon: React.FC<IconProps> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" className={className}><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" /></svg>
);

export const RocketIcon: React.FC<IconProps> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" className={className} strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.63 2.18c-2.35 0-4.59 1.78-4.59 5.84 0 2.84 1.03 5.05 2.2 6.37a1.5 1.5 0 01-2.2 2.2L6 18.5a1.5 1.5 0 01-2.12 0L2.12 16.7a1.5 1.5 0 010-2.12l1.3-1.28a1.5 1.5 0 012.2 2.2l-1.3 1.28c.52.52.88 1.13 1.13 1.75l1.4-1.4a1.5 1.5 0 012.12 0l2.12 2.12a1.5 1.5 0 010 2.12l-1.28 1.3c.62.25 1.23.6 1.75 1.13l1.28-1.3a1.5 1.5 0 012.2 2.2l-1.28 1.3a1.5 1.5 0 010 2.12l2.12 2.12a1.5 1.5 0 012.12 0l1.28-1.3a1.5 1.5 0 012.2-2.2l1.28 1.3a1.5 1.5 0 010 2.12L19.5 21a1.5 1.5 0 01-2.12 0l-1.28-1.3c-.62.25-1.23.6-1.75 1.13l1.28 1.3a1.5 1.5 0 010 2.12L13.12 21a1.5 1.5 0 01-2.12 0l-1.28-1.3a6 6 0 01-7.38-5.84m5.84-2.58A14.98 14.98 0 0015.59 14.37" /></svg>
);

export const CloseIcon: React.FC<IconProps> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" className={className}><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
);
