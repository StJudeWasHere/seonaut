# SEOnaut Translation System

This document explains how SEOnaut's how translation files are organized, and how to add or modify your own translations.

---

## Introduction

SEOnaut supports multiple languages for its user interface. The default language is English (`en`), but additional languages can be added following the guidelines below.

The application language is configured in the main configuration file (`config`) using the `language` parameter under the `[UI]` section:

```ini
[UI]
language = "en"  # ISO 639-1 language code (e.g., "es", "fr", "fa")
```

---

## Translation Files Structure

Translation files are located in the **`translations/`** directory. Each language has its own file named using its ISO 639-1 code:

```
translations/
├── translation.en.yaml  # English (default language)
├── translation.es.yaml  # Spanish
├── translation.fa.yaml  # Persian
└── ...
```

> **Note for developers**: When adding a new language, ensure the language code matches the `language` configuration value and follows the `translation.[code].yaml` format.

---

## Translation File Format

Translation files use **YAML** format and are organized into **contextual sections**. Each file contains:

- **Keys**: Unique identifiers that **must not be translated** (e.g., `SIGN_IN`, `DELETE_ACCOUNT`).
- **Values**: The actual text displayed in the UI, which **must be translated**.
- **Comments**: Explanations for translators and developers (in English).

### Example Structure:

```yaml
# =============================================
# CONTEXT: Sign in page
# =============================================
SIGN_IN: Sign in to your account
SIGIN_ERROR: The email address or password is not valid.
NEW_USER: New to SEOnaut?
```

### Key Components:

1. **Contextual Sections**:
   Each section groups related texts by page or functionality (e.g., `Sign in page`, `Project dashboard`). This helps with the texts context and is useful to provide a more accurate translation.

   ```yaml
   # =============================================
   # CONTEXT: Sign in page
   # =============================================
   SIGN_IN: Sign in to your account
   ```

2. **Keys (Keywords)**:
   - Unique identifiers in **UPPER_SNAKE_CASE** (e.g., `DELETE_ACCOUNT_WARNING`).
   - **Never translate or modify** these, as they are directly referenced in the code.

3. **Values (Texts)**:
   - The actual UI texts that need translation.
   - Translate naturally, adapting to the target language and culture.
   - Some include **placeholders** (e.g., `%1%`, `%2%`) that **must not be translated or removed**.

   ```yaml
   DELETE_PROJECT_MESSAGE: This action will delete the %1% project and all its related data.
   ```

   In this example, the %1% is replaced with the project name in the UI.

4. **Comments**:
   - Provide context on where and how each text is used.
   - Include usage examples and explanations about placeholders.
   - Can be translated, changed or removed. The do not affect the actual translation and are only used to provide context and relevant information.

   ```yaml
   # %1% will be replaced with the project's URL
   DELETE_PROJECT_MESSAGE: This action will delete the %1% project and all its related data.
   ```

---

## Dynamic Parameters (Placeholders)

Some texts include **placeholders** like `%1%` and `%2%`. These are replaced with dynamic values at runtime.

**Important Rules**:

- **Do not translate or remove placeholders** (e.g., `%1%`).
- If a text has multiple placeholders, ensure the translation maintains their relative positions.

**Example**:

```yaml
# Original (en.yaml):
ISSUES_MESSAGE: "%1% critical issues and %2% alerts have been detected on this site."

# Translation (es.yaml):
ISSUES_MESSAGE: "Se han detectado %1% problemas críticos y %2% alertas en este sitio."
```

---

## Adding a New Language

To add support for a new language:

1. **Create a new file** in `translations/` using the ISO 639-1 language code (e.g., `translation.fr.yaml` for French).
2. **Copy the content** from `translation.en.yaml` as a starting point.
3. **Translate all values** (texts), but **keep keys and placeholders** intact.
4. **Test the translation** in the application to ensure no formatting or display issues.

> **Note**: To test out your new translation you must set the language in the `config` file.

---

## Best Practices for Translators

### 1. Terminology Consistency
- Use the same terms for repeated concepts.
- For technical terms (e.g., *canonical*, *hreflang*, *noindex*), refer to the **glossary** or leave them in English if no clear equivalent exists.

### 2. Text Length
- Some texts (e.g., buttons, labels) have **space limitations** in the UI. Avoid overly long translations.

```yaml
# Original (en):
CRAWL_NOW: Crawl Now

# Too long (es):
CRAWL_NOW: Iniciar rastreo ahora mismo

# Better (shorter):
CRAWL_NOW: Rastrear ahora
```

### 3. Pluralization
- If a text has singular/plural versions, ensure both are translated:

```yaml
URL_AFFECTED: 1 URL affected       # Singular
URLS_AFFECTED: "%1% URLs affected"  # Plural
```

### 4. Context Review
- If a text is unclear, check the **comments** in the YAML file or ask for clarification.

---

## Date translation

SEOnaut uses the Go library [`"github.com/goodsign/monday"`](https://"github.com/goodsign/monday") for all date and time translation.

If your language is not supported by this library it will use English as the default language.

---

## Right-to-Left (RTL) Language Support

SEOnaut includes built-in support for right-to-left (RTL) languages such as Arabic, Persian, and Hebrew. When an RTL language is selected, the application automatically adds the `rtl` class to the body applies the necessary CSS adjustments to ensure proper layout and readability.

> **For Translators:** No Additional Steps Required: Simply translate the text as usual. The RTL adjustments are handled automatically by the CSS.

---

## For Developers

### 1. Adding New Texts
If you add new text to the application:

1. Add the key and English text to `translation.en.yaml`.
2. Organize it in the appropriate contextual section.
3. Add comments if the text includes placeholders.
4. **Notify translators** to update other language files.

### 2. Dynamic Placeholders
When using placeholders in code, ensure they are passed in the correct order.

### 3. Translation Testing
- Verify that all placeholders are correctly replaced.
- Ensure no translations are missing for any key.

---

## Collaboration and Support

- **Report issues**: Open an *issue* in the repository with details about:
  - The problematic key.
  - The affected language.
  - A screenshot if possible.

---

## Example Translation File

Below is a snippet from the `translation.es.yaml` file:

```yaml
# =============================================
# CONTEXT: Sign in page
# =============================================
SIGN_IN: Inicia sesión en tu cuenta
SIGIN_ERROR: El correo electrónico o la contraseña no son válidos.
NEW_USER: ¿Nuevo en SEOnaut?

# =============================================
# CONTEXT: Project dashboard page
# =============================================
DASHBOARD: Panel de control
ISSUES_MESSAGE: "Se han detectado %1% problemas críticos y %2% alertas en este sitio."
CRAWL_HISTORY: Historial de rastreos
```

---

## Final Notes

- **Default Language**: English (`en`) is the reference language.
- **Updates**: This document will be updated when significant changes are made.
- **Questions**: Open an *issue* in the repository if you need help.
