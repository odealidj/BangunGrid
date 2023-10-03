'use client';

import React, { useState } from 'react';
import { X, Building2, User, Key, CheckCircle, AlertCircle } from 'lucide-react';
import { apiService } from '@/services/api';
import { User as UserType } from '@/types';

interface ContractorPortalModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccessAuth: (user: UserType, token: string) => void;
}

export default function ContractorPortalModal({
  isOpen,
  onClose,
  onSuccessAuth
}: ContractorPortalModalProps) {
  const [activeTab, setActiveTab] = useState<'login' | 'register'>('login');
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Form states
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [companyName, setCompanyName] = useState('');
  const [taxId, setTaxId] = useState('');
  const [businessRegNum, setBusinessRegNum] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg(null);

    try {
      if (activeTab === 'login') {
        const result = await apiService.login(email, password);
        localStorage.setItem('bangungrid_token', result.tokens.access_token);
        onSuccessAuth(result.user, result.tokens.access_token);
        onClose();
      } else {
        const result = await apiService.registerContractor({
          name,
          email,
          phone,
          password,
          company_name: companyName,
          tax_id: taxId,
          business_reg_num: businessRegNum
        });
        localStorage.setItem('bangungrid_token', result.tokens.access_token);
        onSuccessAuth(result.user, result.tokens.access_token);
        onClose();
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Terjadi kesalahan autentikasi');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <div>
            <h3 style={{ fontSize: '1.25rem', fontWeight: 800 }}>
              {activeTab === 'login' ? 'Masuk ke Akun BangunGrid' : 'Pendaftaran Kontraktor Proyek (B2B)'}
            </h3>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
              Verifikasi token JWT RSA-256 via KrakenD Gateway
            </span>
          </div>
          <button onClick={onClose} style={{ background: 'transparent', color: 'var(--text-muted)' }}>
            <X size={20} />
          </button>
        </div>

        {/* Tab Switcher */}
        <div className="tab-group">
          <button
            type="button"
            className={`tab-btn ${activeTab === 'login' ? 'active' : ''}`}
            onClick={() => { setActiveTab('login'); setErrorMsg(null); }}
          >
            Masuk (Login)
          </button>
          <button
            type="button"
            className={`tab-btn ${activeTab === 'register' ? 'active' : ''}`}
            onClick={() => { setActiveTab('register'); setErrorMsg(null); }}
          >
            Registrasi B2B Baru
          </button>
        </div>

        {errorMsg && (
          <div style={{ padding: '0.75rem', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--rose-alert)', borderRadius: 'var(--radius-md)', color: '#fca5a5', fontSize: '0.85rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <AlertCircle size={16} />
            <span>{errorMsg}</span>
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {activeTab === 'register' && (
            <>
              <div className="form-group">
                <label className="form-label">Nama Lengkap Penanggung Jawab</label>
                <input
                  type="text"
                  required
                  className="form-input"
                  placeholder="Budi Santoso"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Nomor WhatsApp / HP</label>
                <input
                  type="tel"
                  required
                  className="form-input"
                  placeholder="081299998888"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Nama Perusahaan / PT / CV Kontraktor</label>
                <input
                  type="text"
                  required
                  className="form-input"
                  placeholder="PT Bangun Jaya Konstruksi"
                  value={companyName}
                  onChange={(e) => setCompanyName(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">NPWP Perusahaan</label>
                <input
                  type="text"
                  required
                  className="form-input"
                  placeholder="01.234.567.8-901.000"
                  value={taxId}
                  onChange={(e) => setTaxId(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Nomor Induk Berusaha (NIB)</label>
                <input
                  type="text"
                  required
                  className="form-input"
                  placeholder="9120001234567"
                  value={businessRegNum}
                  onChange={(e) => setBusinessRegNum(e.target.value)}
                />
              </div>
            </>
          )}

          <div className="form-group">
            <label className="form-label">Email Akun</label>
            <input
              type="email"
              required
              className="form-input"
              placeholder="kontraktor@perusahaan.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Kata Sandi (Password)</label>
            <input
              type="password"
              required
              className="form-input"
              placeholder="Minimal 8 karakter"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn btn-primary"
            style={{ width: '100%', marginTop: '1rem' }}
          >
            {loading ? 'Menghubungi KrakenD Gateway...' : activeTab === 'login' ? 'Masuk ke Portal' : 'Daftarkan Kontraktor B2B'}
          </button>
        </form>
      </div>
    </div>
  );
}
