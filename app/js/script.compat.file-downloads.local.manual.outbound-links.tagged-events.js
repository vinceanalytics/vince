!function(){"use strict";var i=window.location,o=window.document,p=o.getElementById("plausible"),l=p.getAttribute("data-api")||(f=(f=p).src.split("/"),n=f[0],f=f[2],n+"//"+f+"/api/event");function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return a=t,(r="localStorage flag")&&console.warn("Ignoring Event: "+r),void(a&&a.callback&&a.callback())}catch(e){}var a,r={},n=(r.n=e,r.u=t&&t.u?t.u:i.href,r.d=p.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),new XMLHttpRequest);n.open("POST",l,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(r)),n.onreadystatechange=function(){4===n.readyState&&t&&t.callback&&t.callback({status:n.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a=0;a<t.length;a++)e.apply(this,t[a]);function s(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var u=1;function r(e){if("auxclick"!==e.type||e.button===u){var t,a,r=function(e){for(;e&&(void 0===e.tagName||!s(e)||!e.href);)e=e.parentNode;return e}(e.target),n=r&&r.href&&r.href.split("?")[0];if(!function e(t,a){if(!t||g<a)return!1;if(w(t))return!0;return e(t.parentNode,a+1)}(r,0))return(t=r)&&t.href&&t.host&&t.host!==i.host?c(e,r,{name:"Outbound Link: Click",props:{url:r.href}}):(t=n)&&(a=t.split(".").pop(),m.some(function(e){return e===a}))?c(e,r,{name:"File Download",props:{url:n}}):void 0}}function c(e,t,a){var r,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(r={props:a.props},plausible(a.name,r)):(r={props:a.props,callback:i},plausible(a.name,r),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",r),o.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],f=p.getAttribute("file-types"),d=p.getAttribute("add-file-types"),m=f&&f.split(",")||d&&d.split(",").concat(n)||n;function v(e){var e=w(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var r=0;r<a.length;r++){var n,i=a.item(r).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}var g=3;function b(e){if("auxclick"!==e.type||e.button===u){for(var t,a,r,n,i=e.target,o=0;o<=g&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;s(i)&&(t=i),w(i)&&(a=i),i=i.parentNode}a&&(n=v(a),t?(n.props.url=t.href,c(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function w(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,a=e.target,r=v(a);function n(){t||(t=!0,a.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:r.props,callback:n},plausible(r.name,e))}),o.addEventListener("click",b),o.addEventListener("auxclick",b)}();