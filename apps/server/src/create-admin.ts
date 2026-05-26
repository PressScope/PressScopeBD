import { auth } from "@PressScopeBd/auth";
import { env } from "@PressScopeBd/env/server";

async function createAdmin() {
  try {
    const result = await auth.api.createUser({
      body: {
        email: env.ADMIN_EMAIL,
        password: env.ADMIN_PASSWORD,
        name: env.ADMIN_NAME,
        role: "admin",
      },
    });

    if (result.user.email === env.ADMIN_EMAIL) {
      console.log("Admin user created successfully:", result);
      process.exit(0); // Exits the Node.js process
    } else {
      console.error("Failed to create admin user:", result);
      process.exit(1); // Exits the Node.js process
    }
  } catch (error) {
    console.error("Error creating admin user:", error);
    process.exit(1); // Exits the Node.js process
  }
}

createAdmin();
