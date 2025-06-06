interface User {
	id: number;
	telegram_id: number;
	is_active: boolean;
	is_admin: boolean;
	contact?: Contact;
	created_at: string;
} 