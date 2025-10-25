import { Icon } from "@iconify/react";

interface IconProps {
    size?: number;
    icon: string;
}

export default function Iconify({ icon, size }: IconProps) {
    return (
        <Icon icon={icon} fontSize={size} />
    );
}