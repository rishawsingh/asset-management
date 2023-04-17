ALTER TYPE asset_type ADD VALUE'mobile' AFTER 'pen drive';
ALTER TYPE asset_type ADD VALUE'sim' AFTER 'mobile';
CREATE TABLE IF NOT EXISTS mobile_specifications(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID REFERENCES assets(id) NOT NULL,
    os_type  TEXT,
    imei_1   TEXT,
    imei_2   TEXT,
    ram      TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE ,
    archived_at TIMESTAMP WITH TIME ZONE
);
CREATE TABLE IF NOT EXISTS sim_specifications(
                                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                    asset_id UUID REFERENCES assets(id) NOT NULL,
                                                    sim_no  TEXT,
                                                    phone_no   TEXT,
                                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                                    updated_at TIMESTAMP WITH TIME ZONE ,
                                                    archived_at TIMESTAMP WITH TIME ZONE
);