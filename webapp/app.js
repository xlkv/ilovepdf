// Telegram WebApp Initialization
const tg = window.Telegram ? window.Telegram.WebApp : null;
if (tg) {
  tg.ready();
  tg.expand();
}

// PDF.js worker setup
pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js';

let pdfDocBytes = null;
let pagesData = []; // Array of { originalIndex, rotation, id }

const dropZone = document.getElementById('drop-zone');
const fileInput = document.getElementById('file-input');
const editorSection = document.getElementById('editor-section');
const pagesGrid = document.getElementById('pages-grid');
const loadingOverlay = document.getElementById('loading-overlay');
const loadingText = document.getElementById('loading-text');
const pageCountBadge = document.getElementById('page-count-badge');

// Initialize SortableJS on pages grid
new Sortable(pagesGrid, {
  animation: 150,
  ghostClass: 'sortable-ghost',
  onEnd: () => {
    updatePageNumbersUI();
  }
});

// Event Listeners
fileInput.addEventListener('change', handleFileSelect);

dropZone.addEventListener('dragover', (e) => {
  e.preventDefault();
  dropZone.classList.add('hover');
});

dropZone.addEventListener('dragleave', () => {
  dropZone.classList.remove('hover');
});

dropZone.addEventListener('drop', (e) => {
  e.preventDefault();
  dropZone.classList.remove('hover');
  if (e.dataTransfer.files && e.dataTransfer.files[0]) {
    loadPDFFile(e.dataTransfer.files[0]);
  }
});

document.getElementById('btn-reset').addEventListener('click', () => {
  pdfDocBytes = null;
  pagesData = [];
  pagesGrid.innerHTML = '';
  editorSection.classList.add('hidden');
  dropZone.classList.remove('hidden');
  fileInput.value = '';
});

document.getElementById('btn-rotate-all').addEventListener('click', () => {
  const cards = pagesGrid.querySelectorAll('.page-card');
  cards.forEach(card => {
    let currentRot = parseInt(card.dataset.rotation || '0');
    currentRot = (currentRot + 90) % 360;
    card.dataset.rotation = currentRot;
    const canvas = card.querySelector('canvas');
    if (canvas) {
      canvas.style.transform = `rotate(${currentRot}deg)`;
    }
  });
});

document.getElementById('btn-export').addEventListener('click', exportProcessedPDF);

function handleFileSelect(e) {
  if (e.target.files && e.dataTransfer ? e.dataTransfer.files[0] : e.target.files[0]) {
    const file = e.target.files[0];
    loadPDFFile(file);
  }
}

async function loadPDFFile(file) {
  if (file.type !== 'application/pdf') {
    alert('Iltimos, faqat PDF fayl yuklang!');
    return;
  }

  showLoading('PDF Sahifalari yuklanmoqda...');

  try {
    const arrayBuffer = await file.arrayBuffer();
    pdfDocBytes = new Uint8Array(arrayBuffer);

    const pdfDoc = await pdfjsLib.getDocument({ data: pdfDocBytes.slice(0) }).promise;
    const numPages = pdfDoc.numPages;

    pagesGrid.innerHTML = '';
    pagesData = [];

    for (let i = 1; i <= numPages; i++) {
      pagesData.push({ id: `page-${i}`, originalIndex: i - 1, rotation: 0 });

      const page = await pdfDoc.getPage(i);
      const viewport = page.getViewport({ scale: 0.3 });

      const card = document.createElement('div');
      card.className = 'page-card';
      card.dataset.pageId = `page-${i}`;
      card.dataset.originalIndex = i - 1;
      card.dataset.rotation = 0;

      const numSpan = document.createElement('div');
      numSpan.className = 'page-number';
      numSpan.textContent = `Sahifa ${i}`;

      const wrapper = document.createElement('div');
      wrapper.className = 'page-canvas-wrapper';

      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      canvas.height = viewport.height;
      canvas.width = viewport.width;

      await page.render({ canvasContext: ctx, viewport: viewport }).promise;

      wrapper.appendChild(canvas);

      const controls = document.createElement('div');
      controls.className = 'page-controls';

      const btnRotate = document.createElement('button');
      btnRotate.className = 'icon-btn';
      btnRotate.innerHTML = '🔄';
      btnRotate.title = 'Burish 90°';
      btnRotate.onclick = (e) => {
        e.stopPropagation();
        let rot = parseInt(card.dataset.rotation || '0');
        rot = (rot + 90) % 360;
        card.dataset.rotation = rot;
        canvas.style.transform = `rotate(${rot}deg)`;
      };

      const btnDelete = document.createElement('button');
      btnDelete.className = 'icon-btn';
      btnDelete.innerHTML = '🗑';
      btnDelete.title = 'O\'chirish';
      btnDelete.onclick = (e) => {
        e.stopPropagation();
        card.remove();
        updatePageNumbersUI();
      };

      controls.appendChild(btnRotate);
      controls.appendChild(btnDelete);

      card.appendChild(numSpan);
      card.appendChild(wrapper);
      card.appendChild(controls);

      pagesGrid.appendChild(card);
    }

    dropZone.classList.add('hidden');
    editorSection.classList.remove('hidden');
    updatePageNumbersUI();
  } catch (err) {
    alert('PDF tayyorlashda xatolik: ' + err.message);
  } finally {
    hideLoading();
  }
}

function updatePageNumbersUI() {
  const cards = pagesGrid.querySelectorAll('.page-card');
  pageCountBadge.textContent = `${cards.length} Sahifa`;
  cards.forEach((card, index) => {
    const badge = card.querySelector('.page-number');
    if (badge) {
      badge.textContent = `Sahifa ${index + 1}`;
    }
  });
}

async function exportProcessedPDF() {
  const cards = pagesGrid.querySelectorAll('.page-card');
  if (cards.length === 0) {
    alert('Kamida 1 ta sahifa qoldiring!');
    return;
  }

  showLoading('Yangi PDF yaratilmoqda...');

  try {
    const { PDFDocument, degrees } = PDFLib;
    const srcDoc = await PDFDocument.load(pdfDocBytes);
    const newDoc = await PDFDocument.create();

    for (const card of cards) {
      const origIndex = parseInt(card.dataset.originalIndex);
      const rotation = parseInt(card.dataset.rotation || '0');

      const [copiedPage] = await newDoc.copyPages(srcDoc, [origIndex]);
      if (rotation !== 0) {
        const currentRot = copiedPage.getRotation().angle;
        copiedPage.setRotation(degrees((currentRot + rotation) % 360));
      }
      newDoc.addPage(copiedPage);
    }

    const modifiedPdfBytes = await newDoc.save();
    const blob = new Blob([modifiedPdfBytes], { type: 'application/pdf' });

    // Download locally
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = 'ilovepdf_organized.pdf';
    link.click();

    if (tg) {
      tg.sendData(JSON.stringify({ action: 'pdf_organized', count: cards.length }));
      tg.close();
    }
  } catch (err) {
    alert('Saqlashda xatolik: ' + err.message);
  } finally {
    hideLoading();
  }
}

function showLoading(msg) {
  loadingText.textContent = msg;
  loadingOverlay.classList.remove('hidden');
}

function hideLoading() {
  loadingOverlay.classList.add('hidden');
}
