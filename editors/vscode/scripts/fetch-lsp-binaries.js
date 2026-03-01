const fs = require('fs');
const path = require('path');
const https = require('https');
const zlib = require('zlib');
const tar = require('tar');

const { kukichaVersion: KUKICHA_VERSION } = require('../package.json');
const GITHUB_RELEASES_URL = 'https://github.com/duber000/kukicha/releases/download';

const PLATFORMS = {
    'win32-x64': {
        asset: `kukicha-windows-amd64.tar.gz`,
        dir: 'win32'
    },
    'darwin-x64': {
        asset: `kukicha-darwin-amd64.tar.gz`,
        dir: 'darwin-x64'
    },
    'darwin-arm64': {
        asset: `kukicha-darwin-arm64.tar.gz`,
        dir: 'darwin-arm64'
    },
    'linux-x64': {
        asset: `kukicha-linux-amd64.tar.gz`,
        dir: 'linux'
    }
};

async function downloadAndExtract(url, destDir) {
    return new Promise((resolve, reject) => {
        https.get(url, (response) => {
            if (response.statusCode === 301 || response.statusCode === 302) {
                downloadAndExtract(response.headers.location, destDir).then(resolve).catch(reject);
                return;
            }

            if (response.statusCode !== 200) {
                reject(new Error(`Failed to download: ${response.statusCode}`));
                return;
            }

            // All release assets are .tar.gz — always gunzip then extract.
            const gunzip = zlib.createGunzip();
            const extractor = tar.x({ C: destDir, strip: 1 });

            response.pipe(gunzip).pipe(extractor);
            extractor.on('finish', resolve);
            extractor.on('error', reject);
            gunzip.on('error', reject);
        }).on('error', reject);
    });
}

async function fetchBinary(platform) {
    const config = PLATFORMS[platform];
    if (!config) {
        console.log(`Skipping unknown platform: ${platform}`);
        return;
    }

    const url = `${GITHUB_RELEASES_URL}/v${KUKICHA_VERSION}/${config.asset}`;
    const destDir = path.join(__dirname, '..', 'binaries', config.dir);

    console.log(`Fetching ${platform}: ${url}`);

    fs.mkdirSync(destDir, { recursive: true });

    try {
        await downloadAndExtract(url, destDir);

        // Ensure the binary has the correct name and is executable.
        const binaryName = platform.startsWith('win32') ? 'kukicha-lsp.exe' : 'kukicha-lsp';
        const destBinary = path.join(destDir, binaryName);
        if (fs.existsSync(destBinary)) {
            fs.chmodSync(destBinary, 0o755);
        }

        console.log(`✓ Downloaded ${platform}`);
    } catch (error) {
        console.error(`✗ Failed to download ${platform}: ${error.message}`);
    }
}

async function main() {
    const platforms = process.argv.length > 2 ? 
        process.argv.slice(2) : 
        Object.keys(PLATFORMS);
    
    console.log(`Fetching kukicha-lsp v${KUKICHA_VERSION} binaries...`);
    
    for (const platform of platforms) {
        await fetchBinary(platform);
    }
    
    console.log('\nDone!');
}

main().catch(console.error);
