!function(){"use strict";var r=window.location,i=window.document,o=i.currentScript,l=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function p(t,e){t&&console.warn("Ignoring Event: "+t),e&&e.callback&&e.callback()}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return p("localhost",e);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return p(null,e);try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",e)}catch(t){}var a={},n=(a.n=t,a.u=e&&e.u?e.u:r.href,a.d=o.getAttribute("data-domain"),a.r=i.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),new XMLHttpRequest);n.open("POST",l,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(a)),n.onreadystatechange=function(){4===n.readyState&&e&&e.callback&&e.callback({status:n.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a=0;a<e.length;a++)t.apply(this,e[a]);var s=1;function n(t){var e,a,n,r,i,o,l;function p(){r||(r=!0,window.location=n.href)}"auxclick"===t.type&&t.button!==s||(e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(o=a)&&(l=o.split(".").pop(),f.some(function(t){return t===l}))&&(r=!(o={name:"File Download",props:{url:a}}),!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(a=t,n=e)?(i={props:o.props},plausible(o.name,i)):(i={props:o.props,callback:p},plausible(o.name,i),setTimeout(p,5e3),a.preventDefault())))}i.addEventListener("click",n),i.addEventListener("auxclick",n);var c=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],u=o.getAttribute("file-types"),d=o.getAttribute("add-file-types"),f=u&&u.split(",")||d&&d.split(",").concat(c)||c}();