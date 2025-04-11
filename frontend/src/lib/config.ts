interface Config {
  api: {
    classifierUrl: string;
  };
  auth: {
    jwtSecret: string;
  };
  database: {
    url: string;
  };
}

const config: Config = {
  api: {
    classifierUrl: process.env.CLASSIFIER_URL || 'http://localhost:8080',
  },
  auth: {
    jwtSecret: process.env.JWT_SECRET || 'your-secret-key',
  },
  database: {
    url: process.env.DATABASE_URL || '',
  },
};

export default config;
